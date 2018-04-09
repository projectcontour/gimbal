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
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
)

const (
	actionAdd    = "add"
	actionUpdate = "update"
	actionDelete = "delete"
)

// Queue syncs resources with the Gimbal cluster by working through a queue of
// actions that must be performed against services and endpoints.
type Queue struct {
	Logger      *logrus.Logger
	KubeClient  kubernetes.Interface
	Workqueue   workqueue.RateLimitingInterface
	Threadiness int
}

// Action that is added to the queue for processing
type Action interface {
	Sync(kubernetes.Interface) error
}

// Enqueue adds a new resource action to the worker queue
func (sq *Queue) Enqueue(action Action) {
	sq.Workqueue.AddRateLimited(action)
}

// Run starts the queue workers. It blocks until the stopCh is closed.
func (sq *Queue) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer sq.Workqueue.ShutDown()

	sq.Logger.Infof("Starting queue workers")
	// Launch two workers to process Foo resources
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

	// We wrap this block in a func so we can defer sq.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished processing
		// this item. We also must remember to call Forget if we do not want
		// this work item being re-queued. For example, we do not call Forget if
		// a transient error occurs, instead the item is put back on the
		// workqueue and attempted again after a back-off period.
		defer sq.Workqueue.Done(obj)

		action, ok := obj.(Action)
		if !ok {
			sq.Workqueue.Forget(obj)
			return fmt.Errorf("ignoring unknown item of type %T in queue", obj)
		}

		err := action.Sync(sq.KubeClient)
		if err != nil {
			return err
		}
		sq.Logger.Infof("Successfully handled: %s", action)

		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		sq.Workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		sq.Logger.Error(err)
		return true
	}
	return true
}
