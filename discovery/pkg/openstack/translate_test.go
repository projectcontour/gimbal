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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/pools"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestKubeServices(t *testing.T) {
	tests := []struct {
		name        string
		backendName string
		tenantName  string
		lbs         []loadbalancers.LoadBalancer
		expected    []v1.Service
	}{
		{
			name:        "unnamed lb",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", ""),
			},
			expected: []v1.Service{
				service("finance", "us-east-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": ""},
					nil),
			},
		},
		{
			name:        "named lb",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks"),
			},
			expected: []v1.Service{
				service("finance", "us-east-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					nil),
			},
		},
		{
			name:        "lb with long name",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "very-long-openstack-load-balancer-name-that-is-longer-than-the-limit"),
			},
			expected: []v1.Service{
				service("finance", "us-east-very-long-openstack-load-951f9d",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "very-long-openstack-load-balancer-name-that-is-longer-tha951f9d",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "very-long-openstack-load-balancer-name-that-is-longer-tha80b28c"},
					nil),
			},
		},
		{
			name:        "lb with name that begins with a number",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "1234-stocks"),
			},
			expected: []v1.Service{
				service("finance", "us-east-1234-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "1234-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "1234-stocks"},
					nil),
			},
		},
		{
			name:        "lb with name that contains uppercase letters",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "1234-STOCKS"),
			},
			expected: []v1.Service{
				service("finance", "us-east-1234-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "1234-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "1234-STOCKS"},
					nil),
			},
		},
		{
			name:        "long cluster name, normal lb name",
			backendName: "cluster-name-that-is-definitely-too-long-to-be-useful",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "nginx"),
			},
			expected: []v1.Service{
				service("finance", "cluster-name-that-is-defib224b3-nginx-5a5c3d9e-e679-43ec-e1c9a7",
					map[string]string{
						"gimbal.heptio.com/backend":            "cluster-name-that-is-definitely-too-long-to-be-useful",
						"gimbal.heptio.com/service":            "nginx-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "nginx"},
					nil),
			},
		},
		{
			name:        "long lb name and long cluster name",
			backendName: "cluster-name-that-is-definitely-too-long-to-be-useful",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "very-long-openstack-load-balancer-name-that-is-longer-than-the-limit"),
			},
			expected: []v1.Service{
				service("finance", "cluster-name-that-is-defib224b3-very-long-openstack-load-951f9d",
					map[string]string{
						"gimbal.heptio.com/backend":            "cluster-name-that-is-definitely-too-long-to-be-useful",
						"gimbal.heptio.com/service":            "very-long-openstack-load-balancer-name-that-is-longer-tha951f9d",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "very-long-openstack-load-balancer-name-that-is-longer-tha80b28c"},
					nil),
			},
		},
		{
			name:        "named lb with one listener",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("ls-1", "http", "tcp", "pool-1", 80)),
			},
			expected: []v1.Service{
				service("finance", "us-east-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.ServicePort{
						{
							Name:       "port-80",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
					}),
			},
		},
		{
			name:        "one named lb, one listener",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("ls-1", "http", "tcp", "pool-1", 80)),
			},
			expected: []v1.Service{
				service("finance", "us-east-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.ServicePort{
						{
							Name:       "port-80",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
					}),
			},
		},
		{
			name:        "named lb, two listeners",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("ls-1", "http", "tcp", "pool-1", 80), listener("ls-1", "https", "tcp", "pool-1", 443)),
			},
			expected: []v1.Service{
				service("finance", "us-east-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.ServicePort{
						{
							Name:       "port-80",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
						{
							Name:       "port-443",
							Port:       443,
							TargetPort: intstr.FromInt(443),
							Protocol:   v1.ProtocolTCP,
						},
					}),
			},
		},
		{
			name:        "unnammed lb, two listeners",
			tenantName:  "finance",
			backendName: "us-east",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "", listener("ls-1", "http", "tcp", "pool-1", 80), listener("ls-1", "https", "tcp", "pool-1", 443)),
			},
			expected: []v1.Service{
				service("finance", "us-east-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": ""},
					[]v1.ServicePort{
						{
							Name:       "port-80",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   v1.ProtocolTCP,
						},
						{
							Name:       "port-443",
							Port:       443,
							TargetPort: intstr.FromInt(443),
							Protocol:   v1.ProtocolTCP,
						},
					}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := kubeServices(tc.backendName, tc.tenantName, tc.lbs)
			assert.Equal(t, tc.expected, got)
			assert.Len(t, got, len(tc.lbs))
		})
	}
}

func TestKubeEndpoints(t *testing.T) {
	tests := []struct {
		name        string
		tenantName  string
		backendName string
		lbs         []loadbalancers.LoadBalancer
		pools       []pools.Pool
		expected    []v1.Endpoints
	}{
		{
			name:        "single listener",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("970fd223-4684-4c50-bfa2-738d6dda096f", "http", "tcp", "pool-1", 80)),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e", poolmember("10.0.0.1", 8080)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "us-east-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}},
							Ports:     []v1.EndpointPort{{Name: "port-80", Port: 8080, Protocol: v1.ProtocolTCP}},
						},
					}),
			},
		},
		{
			name:        "single listener, long cluster name",
			backendName: "cluster-name-that-is-definitely-too-long-to-be-useful-in-any-shape-or-form",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("970fd223-4684-4c50-bfa2-738d6dda096f", "http", "tcp", "pool-1", 80)),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e", poolmember("10.0.0.1", 8080)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "cluster-name-that-is-defi7afb09-stocks-5a5c3d9e-e679-43ec549cd6",
					map[string]string{
						"gimbal.heptio.com/backend":            "cluster-name-that-is-definitely-too-long-to-be-useful-in-7afb09",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}},
							Ports:     []v1.EndpointPort{{Name: "port-80", Port: 8080, Protocol: v1.ProtocolTCP}},
						},
					}),
			},
		},
		{
			name:        "single listener, long load balancer name",
			backendName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "very-long-openstack-load-balancer-name-that-is-longer-than-the-limit", listener("970fd223-4684-4c50-bfa2-738d6dda096f", "http", "tcp", "pool-1", 80)),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e", poolmember("10.0.0.1", 8080)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "us-east-very-long-openstack-load-951f9d",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "very-long-openstack-load-balancer-name-that-is-longer-tha951f9d",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "very-long-openstack-load-balancer-name-that-is-longer-tha80b28c"},
					[]v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}},
							Ports:     []v1.EndpointPort{{Name: "port-80", Port: 8080, Protocol: v1.ProtocolTCP}},
						},
					}),
			},
		},
		{
			name:        "multiple listeners",
			tenantName:  "finance",
			backendName: "us-east",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("loadbalancer-1", "stocks", listener("listener-1", "http", "tcp", "pool-1", 80), listener("listener-2", "https", "tcp", "pool-2", 443)),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "loadbalancer-1", poolmember("10.0.0.1", 8080), poolmember("10.0.0.2", 80), poolmember("10.0.0.3", 80), poolmember("10.0.0.4", 8080)),
				pool("pool-2", "HTTP", "loadbalancer-1", poolmember("10.0.0.5", 443), poolmember("10.0.0.6", 443), poolmember("10.0.0.7", 8443)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "us-east-stocks-loadbalancer-1",
					map[string]string{
						"gimbal.heptio.com/backend":            "us-east",
						"gimbal.heptio.com/service":            "stocks-loadbalancer-1",
						"gimbal.heptio.com/load-balancer-id":   "loadbalancer-1",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}, {IP: "10.0.0.4"}},
							Ports:     []v1.EndpointPort{{Name: "port-80", Port: 8080, Protocol: v1.ProtocolTCP}},
						},
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.2"}, {IP: "10.0.0.3"}},
							Ports:     []v1.EndpointPort{{Name: "port-80", Port: 80, Protocol: v1.ProtocolTCP}},
						},
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.5"}, {IP: "10.0.0.6"}},
							Ports:     []v1.EndpointPort{{Name: "port-443", Port: 443, Protocol: v1.ProtocolTCP}},
						},
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.7"}},
							Ports:     []v1.EndpointPort{{Name: "port-443", Port: 8443, Protocol: v1.ProtocolTCP}},
						},
					}),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotReturn := kubeEndpoints(tc.backendName, tc.tenantName, tc.lbs, tc.pools)
			// Cannot use assert.Equal on the structs as the order of subsets is undetermined.
			var got []Endpoints
			for _, v := range gotReturn {
				got = append(got, v)
			}

			for i := range tc.expected {
				assert.Equal(t, tc.expected[i].Namespace, got[i].endpoints.Namespace)
				assert.Equal(t, tc.expected[i].Name, got[i].endpoints.Name)
				assert.Equal(t, tc.expected[i].Labels, got[i].endpoints.Labels)
				assert.ElementsMatch(t, tc.expected[i].Subsets, got[i].endpoints.Subsets)
			}
		})
	}
}

func service(namespace, name string, labels map[string]string, ports []v1.ServicePort) v1.Service {
	return v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "None",
			Ports:     ports,
		},
	}
}

func endpoints(namespace, name string, labels map[string]string, subsets []v1.EndpointSubset) v1.Endpoints {
	return v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Subsets: subsets,
	}
}

func loadbalancer(id string, name string, listeners ...listeners.Listener) loadbalancers.LoadBalancer {
	return loadbalancers.LoadBalancer{
		ID:        id,
		Name:      name,
		Listeners: listeners,
	}
}

func listener(id, name, protocol, poolID string, port int) listeners.Listener {
	return listeners.Listener{
		ID:            id,
		Name:          name,
		ProtocolPort:  port,
		Protocol:      protocol,
		DefaultPoolID: poolID,
	}
}

func pool(id, protocol, loadBalancerID string, members ...pools.Member) pools.Pool {
	return pools.Pool{
		ID:            id,
		Protocol:      protocol,
		Loadbalancers: []pools.LoadBalancerID{{ID: loadBalancerID}},
		Members:       members,
	}
}

func poolmember(address string, port int) pools.Member {
	return pools.Member{Address: address, ProtocolPort: port}
}
