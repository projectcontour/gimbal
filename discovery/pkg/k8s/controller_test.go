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

	"github.com/heptio/gimbal/discovery/pkg/sync"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
)

var serviceTests = []struct {
	name     string
	service  *v1.Service
	expected int
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
		expected: 1,
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
		expected: 0,
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
		expected: 0,
	},
}

var endpointTests = []struct {
	name     string
	endpoint *v1.Endpoints
	expected int
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
		expected: 1,
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
		expected: 0,
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
		expected: 0,
	},
}

func TestAddServiceQueue(t *testing.T) {
	for _, tc := range serviceTests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Controller{
				Logger: logrus.New(),
				syncqueue: sync.Queue{
					Logger:      logrus.New(),
					Workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "syncqueue"),
					Threadiness: 1,
					ClusterName: "test",
				},
				clusterName: "test",
			}

			c.addService(tc.service)
			time.Sleep(1 * time.Second) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestUpdateServiceQueue(t *testing.T) {
	for _, tc := range serviceTests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Controller{
				Logger: logrus.New(),
				syncqueue: sync.Queue{
					Logger:      logrus.New(),
					Workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "syncqueue"),
					Threadiness: 1,
					ClusterName: "test",
				},
				clusterName: "test",
			}

			c.updateService(tc.service)
			time.Sleep(1 * time.Second) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestDeleteServiceQueue(t *testing.T) {
	for _, tc := range serviceTests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Controller{
				Logger: logrus.New(),
				syncqueue: sync.Queue{
					Logger:      logrus.New(),
					Workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "syncqueue"),
					Threadiness: 1,
					ClusterName: "test",
				},
				clusterName: "test",
			}

			c.deleteService(tc.service)
			time.Sleep(1 * time.Second) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestAddEndpointsQueue(t *testing.T) {
	for _, tc := range endpointTests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Controller{
				Logger: logrus.New(),
				syncqueue: sync.Queue{
					Logger:      logrus.New(),
					Workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "syncqueue"),
					Threadiness: 1,
					ClusterName: "test",
				},
				clusterName: "test",
			}

			c.addEndpoints(tc.endpoint)
			time.Sleep(1 * time.Second) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestUpdateEndpointsQueue(t *testing.T) {
	for _, tc := range endpointTests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Controller{
				Logger: logrus.New(),
				syncqueue: sync.Queue{
					Logger:      logrus.New(),
					Workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "syncqueue"),
					Threadiness: 1,
					ClusterName: "test",
				},
				clusterName: "test",
			}

			c.updateEndpoints(tc.endpoint)
			time.Sleep(1 * time.Second) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
func TestDeleteEndpointsQueue(t *testing.T) {
	for _, tc := range endpointTests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Controller{
				Logger: logrus.New(),
				syncqueue: sync.Queue{
					Logger:      logrus.New(),
					Workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "syncqueue"),
					Threadiness: 1,
					ClusterName: "test",
				},
				clusterName: "test",
			}

			c.deleteEndpoints(tc.endpoint)
			time.Sleep(1 * time.Second) // Give queue time to process (huh?)
			got := c.syncqueue.Workqueue.Len()
			assert.Equal(t, tc.expected, got)
		})
	}
}
