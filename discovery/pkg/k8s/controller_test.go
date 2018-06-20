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

package k8s

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/heptio/gimbal/discovery/pkg/sync"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/util/workqueue"
)

var serviceTests = []struct {
	name                  string
	service               *v1.Service
	expected              int
	expectedServicesCount int
}{
	{
		name: "service",
		service: &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test", Namespace: "default"},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
		expected:              1,
		expectedServicesCount: 1,
	},
	{
		name: "service into kube-system namespace",
		service: &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test", Namespace: "kube-system"},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
		expected:              0,
		expectedServicesCount: 1,
	},
	{
		name: "kubernetes service",
		service: &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kubernetes", Namespace: "default"},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
		expected:              0,
		expectedServicesCount: 0,
	},
	{
		name: "kubernetes service diff namespace",
		service: &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kubernetes", Namespace: "foo"},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
		expected:              1,
		expectedServicesCount: 1,
	},
}

var endpointTests = []struct {
	name                   string
	endpoint               *v1.Endpoints
	expected               int
	expectedEndpointsCount int
}{
	{
		name: "add endpoint",
		endpoint: &v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test", Namespace: "default",
			},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{{IP: "192.168.0.1"}},
					Ports:     []v1.EndpointPort{{Port: 80}},
				},
			},
		},
		expected:               1,
		expectedEndpointsCount: 1,
	},
	{
		name: "service into kube-system namespace",
		endpoint: &v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test", Namespace: "kube-system"},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{{IP: "192.168.0.1"}},
					Ports:     []v1.EndpointPort{{Port: 80}},
				},
			},
		},
		expected:               0,
		expectedEndpointsCount: 1,
	},
	{
		name: "kubernetes endpoint",
		endpoint: &v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kubernetes", Namespace: "default"},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{{IP: "192.168.0.1"}},
					Ports:     []v1.EndpointPort{{Port: 80}},
				},
			},
		},
		expected:               0,
		expectedEndpointsCount: 1,
	},
	{
		name: "kubernetes endpoint diff namespace",
		endpoint: &v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kubernetes", Namespace: "foo"},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{{IP: "192.168.0.1"}},
					Ports:     []v1.EndpointPort{{Port: 80}},
				},
			},
		},
		expected:               1,
		expectedEndpointsCount: 1,
	},
}

func TestAddServiceQueue(t *testing.T) {
	for _, tc := range serviceTests {
		t.Run(tc.name, func(t *testing.T) {
			c := getDefaultController(localmetrics.NewMetrics("backendtype", "backend"))
			c.addService(tc.service)
			time.Sleep(100 * time.Millisecond) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestUpdateServiceQueue(t *testing.T) {
	for _, tc := range serviceTests {
		t.Run(tc.name, func(t *testing.T) {
			c := getDefaultController(localmetrics.NewMetrics("backendtype", "backend"))
			c.updateService(tc.service)
			time.Sleep(100 * time.Millisecond) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestDeleteServiceQueue(t *testing.T) {
	for _, tc := range serviceTests {
		t.Run(tc.name, func(t *testing.T) {
			c := getDefaultController(localmetrics.NewMetrics("backendtype", "backend"))
			c.deleteService(tc.service)
			time.Sleep(100 * time.Millisecond) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestAddEndpointsQueue(t *testing.T) {
	for _, tc := range endpointTests {
		t.Run(tc.name, func(t *testing.T) {
			c := getDefaultController(localmetrics.NewMetrics("backendtype", "backend"))
			c.addEndpoints(tc.endpoint)
			time.Sleep(100 * time.Millisecond) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestUpdateEndpointsQueue(t *testing.T) {
	for _, tc := range endpointTests {
		t.Run(tc.name, func(t *testing.T) {
			c := getDefaultController(localmetrics.NewMetrics("backendtype", "backend"))
			c.updateEndpoints(tc.endpoint)
			time.Sleep(100 * time.Millisecond) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestDeleteEndpointsQueue(t *testing.T) {
	for _, tc := range endpointTests {
		t.Run(tc.name, func(t *testing.T) {
			c := getDefaultController(localmetrics.NewMetrics("backendtype", "backend"))
			c.deleteEndpoints(tc.endpoint)
			time.Sleep(100 * time.Millisecond) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestEndpointMetrics(t *testing.T) {
	for _, tc := range endpointTests {
		t.Run(tc.name, func(t *testing.T) {

			metrics := localmetrics.NewMetrics("backendtype", "backend")
			metrics.RegisterPrometheus(false)

			c := getDefaultController(metrics)

			c.writeEndpointsMetrics(tc.endpoint)

			gatherers := prometheus.Gatherers{
				metrics.Registry,
				prometheus.DefaultGatherer,
			}

			gathering, err := gatherers.Gather()
			if err != nil {
				t.Fatal(err)
			}

			replicatedEndpoints := float64(-1)
			for _, mf := range gathering {
				if mf.GetName() == localmetrics.DiscovererUpstreamEndpointsGauge {
					replicatedEndpoints = mf.Metric[0].Gauge.GetValue()
				}
			}

			assert.Equal(t, float64(tc.expectedEndpointsCount), replicatedEndpoints)
		})
	}
}
func TestServicesMetrics(t *testing.T) {
	for _, tc := range serviceTests {
		t.Run(tc.name, func(t *testing.T) {

			metrics := localmetrics.NewMetrics("backendtype", "backend")
			metrics.RegisterPrometheus(false)

			client := fake.NewSimpleClientset(tc.service)
			informer := kubeinformers.NewSharedInformerFactory(client, time.Second*0)

			c := &Controller{
				Logger: logrus.New(),
				syncqueue: sync.Queue{
					Logger:      logrus.New(),
					Workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "syncqueue"),
					Threadiness: 1,
					Metrics:     metrics,
				},
				serviceLister:   informer.Core().V1().Services().Lister(),
				endpointsLister: informer.Core().V1().Endpoints().Lister(),
				metrics:         metrics,
			}

			// Call informer before starting!
			informer.Core().V1().Services().Informer()

			stopCh := make(chan struct{})
			informer.Start(stopCh)
			informer.WaitForCacheSync(stopCh)

			c.writeServiceMetrics(tc.service)

			gatherers := prometheus.Gatherers{
				metrics.Registry,
				prometheus.DefaultGatherer,
			}

			gathering, err := gatherers.Gather()
			if err != nil {
				t.Fatal(err)
			}

			replicatedServices := float64(-1)
			for _, mf := range gathering {
				if mf.GetName() == localmetrics.DiscovererUpstreamServicesGauge {
					replicatedServices = mf.Metric[0].Gauge.GetValue()
				}
			}

			assert.Equal(t, float64(tc.expectedServicesCount), replicatedServices)
		})
	}
}

func getDefaultController(metrics localmetrics.DiscovererMetrics) *Controller {
	client := fake.NewSimpleClientset()
	informer := kubeinformers.NewSharedInformerFactory(client, time.Second*0)
	return &Controller{
		Logger: logrus.New(),
		syncqueue: sync.Queue{
			Logger:      logrus.New(),
			Workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "syncqueue"),
			Threadiness: 1,
			Metrics:     metrics,
		},
		serviceLister:   informer.Core().V1().Services().Lister(),
		endpointsLister: informer.Core().V1().Endpoints().Lister(),
		metrics:         metrics,
	}
}
