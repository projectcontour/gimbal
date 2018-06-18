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
	"regexp"
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

	// BackendName is the name of the OpenStack cluster
	BackendName string
	ClusterType string
	// GimbalKubeClient is the client of the Kubernetes cluster where Gimbal is running
	GimbalKubeClient kubernetes.Interface
	// Interval between reconciliation loops
	SyncPeriod time.Duration
	Logger     *logrus.Logger
	syncqueue  sync.Queue

	Metrics localmetrics.DiscovererMetrics
}

// Endpoints represents a v1.Endpoints + upstream name to facilicate metrics
type Endpoints struct {
	endpoints    v1.Endpoints
	upstreamName string
}

// NewReconciler returns an OpenStack reconciler
func NewReconciler(backendName, clusterType string, gimbalKubeClient kubernetes.Interface, syncPeriod time.Duration, lbLister LoadBalancerLister,
	projectLister ProjectLister, log *logrus.Logger, queueWorkers int, metrics localmetrics.DiscovererMetrics) Reconciler {

	return Reconciler{
		BackendName:        backendName,
		GimbalKubeClient:   gimbalKubeClient,
		SyncPeriod:         syncPeriod,
		LoadBalancerLister: lbLister,
		ProjectLister:      projectLister,
		Logger:             log,
		Metrics:            metrics,
		syncqueue:          sync.NewQueue(log, gimbalKubeClient, queueWorkers, metrics),
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
	log.Info("reconciling load balancers")
	// Get all the openstack tenants that must be synced
	projects, err := r.ProjectLister.ListProjects()
	if err != nil {
		r.Metrics.GenericMetricError("ListProjects")
		log.Errorf("error listing OpenStack projects: %v", err)
		return
	}
	for _, project := range projects {
		projectName := project.Name

		// Get load balancers that are defined in the project
		loadbalancers, err := r.ListLoadBalancers(project.ID)
		if err != nil {
			r.Metrics.GenericMetricError("ListLoadBalancers")
			log.Errorf("error reconciling project %q: %v", projectName, err)
			continue
		}

		totalUpstreamServices := len(loadbalancers)
		loadbalancers = r.skipInvalidLoadBalancers(projectName, loadbalancers)
		totalInvalidServices := totalUpstreamServices - len(loadbalancers)

		// Get all pools defined in the project
		pools, err := r.ListPools(project.ID)
		if err != nil {
			r.Metrics.GenericMetricError("ListPools")
			log.Errorf("error reconciling project %q: %v", projectName, err)
			continue
		}

		// Get all services and endpoints that exist in the corresponding namespace
		clusterLabelSelector := fmt.Sprintf("%s=%s", translator.GimbalLabelBackend, r.BackendName)
		currentServices, err := r.GimbalKubeClient.CoreV1().Services(projectName).List(metav1.ListOptions{LabelSelector: clusterLabelSelector})
		if err != nil {
			r.Metrics.GenericMetricError("ListServicesInNamespace")
			log.Errorf("error listing services in namespace %q: %v", projectName, err)
			continue
		}

		currentk8sEndpoints, err := r.GimbalKubeClient.CoreV1().Endpoints(projectName).List(metav1.ListOptions{LabelSelector: clusterLabelSelector})
		if err != nil {
			r.Metrics.GenericMetricError("ListEndpointsInNamespace")
			log.Errorf("error listing endpoints in namespace:%q: %v", projectName, err)
			continue
		}

		// Convert the k8s list to type []Endpoints so make comparison easier
		currentEndpoints := []Endpoints{}
		for _, v := range currentk8sEndpoints.Items {
			currentEndpoints = append(currentEndpoints, Endpoints{endpoints: v, upstreamName: ""})
		}

		// Reconcile current state with desired state
		desiredSvcs := kubeServices(r.BackendName, projectName, loadbalancers)
		r.reconcileSvcs(desiredSvcs, currentServices.Items)

		desiredEndpoints := kubeEndpoints(r.BackendName, projectName, loadbalancers, pools)
		r.reconcileEndpoints(desiredEndpoints, currentEndpoints)

		// Log upstream /invalid services to prometheus
		r.Metrics.DiscovererUpstreamServicesMetric(projectName, totalUpstreamServices)
		r.Metrics.DiscovererInvalidServicesMetric(projectName, totalInvalidServices)

		for _, ep := range desiredEndpoints {
			totalUpstreamEndpoints := sync.SumEndpoints(&ep.endpoints)
			r.Metrics.DiscovererUpstreamEndpointsMetric(projectName, ep.upstreamName, totalUpstreamEndpoints)
		}
	}

	// Log to Prometheus the cycle duration
	r.Metrics.CycleDurationMetric(time.Now().Sub(start))
}

// skip any load balancer that has invalid characters, according to the
// characters allowed by the DNS_LABEL spec in Kubernetes.
func (r *Reconciler) skipInvalidLoadBalancers(projectName string, lbs []loadbalancers.LoadBalancer) []loadbalancers.LoadBalancer {
	// names can include letters, numbers or dashes.
	// names must end with a letter or number.
	validName := regexp.MustCompile("^[-a-zA-Z0-9]+[a-zA-Z0-9]$")
	valid := []loadbalancers.LoadBalancer{}
	for _, lb := range lbs {
		if lb.Name != "" && !validName.MatchString(lb.Name) {
			r.Metrics.GenericMetricError("InvalidLoadBalancerName")
			r.Logger.Warningf("skipping load balancer '%s' in project '%s' as it has an invalid name '%s'", lb.ID, projectName, lb.Name)
			continue
		}
		valid = append(valid, lb)
	}
	return valid
}

func (r *Reconciler) reconcileSvcs(desiredSvcs, currentSvcs []v1.Service) {
	add, up, del := diffServices(desiredSvcs, currentSvcs)
	for _, svc := range add {
		r.syncqueue.Enqueue(sync.AddServiceAction(&svc))
	}
	for _, svc := range up {
		r.syncqueue.Enqueue(sync.UpdateServiceAction(&svc))
	}
	for _, svc := range del {
		r.syncqueue.Enqueue(sync.DeleteServiceAction(&svc))
	}
}

func (r *Reconciler) reconcileEndpoints(desired []Endpoints, current []Endpoints) {
	add, up, del := diffEndpoints(desired, current)
	for _, ep := range add {
		r.syncqueue.Enqueue(sync.AddEndpointsAction(&ep.endpoints, ep.upstreamName))
	}
	for _, ep := range up {
		r.syncqueue.Enqueue(sync.UpdateEndpointsAction(&ep.endpoints, ep.upstreamName))
	}
	for _, ep := range del {
		r.syncqueue.Enqueue(sync.DeleteEndpointsAction(&ep.endpoints, ep.upstreamName))
	}
}
