// Copyright © 2018 Heptio
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sync

import (
	"time"

	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
)

const (
	actionAdd       = "add"
	actionUpdate    = "update"
	actionDelete    = "delete"
	queueMaxRetries = 3
)

// Queue syncs resources with the Gimbal cluster by working through a queue of
// actions that must be performed against services and endpoints.
type Queue struct {
	Logger      *logrus.Logger
	KubeClient  kubernetes.Interface
	Workqueue   workqueue.RateLimitingInterface
	Threadiness int
	Metrics     localmetrics.DiscovererMetrics
}

// NewQueue returns an initialized sync.Queue for syncing resources with a Gimbal cluster.
func NewQueue(logger *logrus.Logger, kubeClient kubernetes.Interface,
	threadiness int, metrics localmetrics.DiscovererMetrics) Queue {
	return Queue{
		KubeClient:  kubeClient,
		Logger:      logger,
		Workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "syncqueue"),
		Threadiness: threadiness,
		Metrics:     metrics,
	}
}

// Action that is added to the queue for processing
type Action interface {
	Sync(kube kubernetes.Interface, logger *logrus.Logger) error
	ObjectMeta() *metav1.ObjectMeta
	SetMetrics(gimbalKubeClient kubernetes.Interface, lm localmetrics.DiscovererMetrics, logger *logrus.Logger)
	SetMetricError(metrics localmetrics.DiscovererMetrics)
	GetActionType() string
}

// Enqueue adds a new resource action to the worker queue
func (sq *Queue) Enqueue(action Action) {
	sq.Workqueue.AddRateLimited(action)
	sq.Metrics.QueueSizeGaugeMetric(sq.Workqueue.Len())
}

// Run starts the queue workers. It blocks until the stopCh is closed.
func (sq *Queue) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer sq.Workqueue.ShutDown()

	sq.Logger.Infof("Starting queue workers")
	// Launch workers to process Action resources
	for i := 0; i < sq.Threadiness; i++ {
		go wait.Until(sq.runWorker, time.Second, stopCh)
	}

	sq.Logger.Infof("Started workers")
	<-stopCh
	sq.Logger.Infof("Shutting down workers")
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (sq *Queue) runWorker() {
	for sq.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (sq *Queue) processNextWorkItem() bool {
	obj, shutdown := sq.Workqueue.Get()

	if shutdown {
		return false
	}

	// Tell the queue that we are done with processing this key. This unblocks
	// the key for other workers.
	defer sq.Workqueue.Done(obj)

	action, ok := obj.(Action)
	if !ok {
		sq.Workqueue.Forget(obj)
		sq.Metrics.QueueSizeGaugeMetric(sq.Workqueue.Len())
		sq.Logger.Errorf("got an unknown item of type %T in the queue", obj)
		return true
	}

	err := action.Sync(sq.KubeClient, sq.Logger)

	// We successfully handled the action, so we can forget the item and keep going.
	if err == nil {
		sq.Workqueue.Forget(obj)
		action.SetMetrics(sq.KubeClient, sq.Metrics, sq.Logger)
		sq.Metrics.QueueSizeGaugeMetric(sq.Workqueue.Len())
		sq.Logger.Infof("Successfully handled: %s", action)
		return true
	}

	// An error ocurred. Set the error metrics.
	action.SetMetricError(sq.Metrics)

	// If there was an error handling the item, we will retry up to
	// queueMaxRetries times.
	numRequeues := sq.Workqueue.NumRequeues(obj)
	if numRequeues < queueMaxRetries {
		sq.Logger.Errorf("Error handling %s: %v. Number of requeues: %d. Requeuing.", action, err, numRequeues)
		sq.Workqueue.AddRateLimited(obj)
		sq.Metrics.QueueSizeGaugeMetric(sq.Workqueue.Len())
		return true
	}

	// We tried `queueMaxRetries` times but still failed. Dropping the item from
	// the queue.
	sq.Workqueue.Forget(obj)
	sq.Logger.Errorf("Dropping %s out of the queue because we failed to handle the item %d times: %v", action, queueMaxRetries, err)
	sq.Metrics.QueueSizeGaugeMetric(sq.Workqueue.Len())
	return true
}
