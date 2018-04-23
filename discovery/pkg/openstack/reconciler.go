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

package openstack

import (
	"fmt"
	"math"
	"time"

	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/pools"
	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/heptio/gimbal/discovery/pkg/sync"
	"github.com/heptio/gimbal/discovery/pkg/translator"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const clusterType = "openstack"

type ProjectLister interface {
	ListProjects() ([]projects.Project, error)
}

type LoadBalancerLister interface {
	ListLoadBalancers(projectID string) ([]loadbalancers.LoadBalancer, error)
	ListPools(projectID string) ([]pools.Pool, error)
}

// The Reconciler connects to an OpenStack cluster and makes sure that the Load
// Balancers defined in the cluster are reflected in the Gimbal Kubernetes
// cluster as Services and Endpoints. The Reconciler runs on a configurable
// interval.
type Reconciler struct {
	LoadBalancerLister
	ProjectLister

	// ClusterName is the name of the OpenStack cluster
	ClusterName string
	// GimbalKubeClient is the client of the Kubernetes cluster where Gimbal is running
	GimbalKubeClient kubernetes.Interface
	// Interval between reconciliation loops
	SyncPeriod time.Duration
	Logger     *logrus.Logger
	syncqueue  sync.Queue

	Metrics localmetrics.DiscovererMetrics
}

// NewReconciler returns an OpenStack reconciler
func NewReconciler(clusterName string, gimbalKubeClient kubernetes.Interface, syncPeriod time.Duration, lbLister LoadBalancerLister,
	projectLister ProjectLister, log *logrus.Logger, queueWorkers int, metrics localmetrics.DiscovererMetrics) Reconciler {
	return Reconciler{
		ClusterName:        clusterName,
		GimbalKubeClient:   gimbalKubeClient,
		SyncPeriod:         syncPeriod,
		LoadBalancerLister: lbLister,
		ProjectLister:      projectLister,
		Logger:             log,
		Metrics:            metrics,
		syncqueue:          sync.NewQueue(log, clusterName, clusterType, gimbalKubeClient, queueWorkers, metrics),
	}
}

// Run starts the reconciler
func (r *Reconciler) Run(stop <-chan struct{}) {
	go r.syncqueue.Run(stop)

	ticker := time.NewTicker(r.SyncPeriod)
	defer ticker.Stop()

	// Perform an initial reconciliation
	r.reconcile()

	// Perform reconciliation on every tick
	for {
		select {
		case <-stop:
			r.Logger.Info("Stopping openstack reconciler")
			return
		case <-ticker.C:
			r.reconcile()
		}
	}
}

func (r *Reconciler) reconcile() {
	// Calculate cycle time
	start := time.Now()

	log := r.Logger
	log.Debugln("reconciling openstack load balancers")
	// Get all the openstack tenants that must be synced
	projects, err := r.ProjectLister.ListProjects()
	if err != nil {
		r.Metrics.GenericMetricError(r.ClusterName, "ListProjects")
		log.Errorf("error listing OpenStack projects: %v", err)
		return
	}
	for _, project := range projects {
		projectName := project.Name

		// Get load balancers that are defined in the project
		loadbalancers, err := r.ListLoadBalancers(project.ID)
		if err != nil {
			r.Metrics.GenericMetricError(r.ClusterName, "ListLoadBalancers")
			log.Errorf("error reconciling project %q: %v", projectName, err)
			continue
		}

		// Get all pools defined in the project
		pools, err := r.ListPools(project.ID)
		if err != nil {
			r.Metrics.GenericMetricError(r.ClusterName, "ListPools")
			log.Errorf("error reconciling project %q: %v", projectName, err)
			continue
		}

		// Get all services and endpoints that exist in the corresponding
		// namespace
		clusterLabelSelector := fmt.Sprintf("%s=%s", translator.GimbalLabelCluster, r.ClusterName)
		currentServices, err := r.GimbalKubeClient.CoreV1().Services(projectName).List(metav1.ListOptions{LabelSelector: clusterLabelSelector})
		if err != nil {
			r.Metrics.GenericMetricError(r.ClusterName, "ListServicesInNamespace")
			log.Errorf("error listing services in namespace %q: %v", projectName, err)
			continue
		}

		currentEndpoints, err := r.GimbalKubeClient.CoreV1().Endpoints(projectName).List(metav1.ListOptions{LabelSelector: clusterLabelSelector})
		if err != nil {
			r.Metrics.GenericMetricError(r.ClusterName, "ListEndpointsInNamespace")
			log.Errorf("error listing endpoints in namespace:%q: %v", projectName, err)
			continue
		}

		// Reconcile current state with desired state
		desiredSvcs := kubeServices(r.ClusterName, projectName, loadbalancers)
		r.reconcileSvcs(desiredSvcs, currentServices.Items)

		desiredEndpoints := kubeEndpoints(r.ClusterName, projectName, loadbalancers, pools)
		r.reconcileEndpoints(desiredEndpoints, currentEndpoints.Items)
	}

	// Log to Prometheus the cycle duration
	r.Metrics.CycleDurationMetric(r.ClusterName, clusterType, math.Floor(time.Now().Sub(start).Seconds()*1e3))
}

func (r *Reconciler) reconcileSvcs(desiredSvcs, currentSvcs []v1.Service) {
	add, up, del := diffServices(desiredSvcs, currentSvcs)
	for _, s := range add {
		svc := s
		r.syncqueue.Enqueue(sync.AddServiceAction(&svc))
	}
	for _, s := range up {
		svc := s
		r.syncqueue.Enqueue(sync.UpdateServiceAction(&svc))
	}
	for _, s := range del {
		svc := s
		r.syncqueue.Enqueue(sync.DeleteServiceAction(&svc))
	}
}

func (r *Reconciler) reconcileEndpoints(desired, current []v1.Endpoints) {
	add, up, del := diffEndpoints(desired, current)
	for _, e := range add {
		ep := e
		r.syncqueue.Enqueue(sync.AddEndpointsAction(&ep))
	}
	for _, e := range up {
		ep := e
		r.syncqueue.Enqueue(sync.UpdateEndpointsAction(&ep))
	}
	for _, e := range del {
		ep := e
		r.syncqueue.Enqueue(sync.DeleteEndpointsAction(&ep))
	}
}
