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
)

func TestKubeServices(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		tenantName  string
		lbs         []loadbalancers.LoadBalancer
		expected    []v1.Service
	}{
		{
			name:        "unnamed lb, no listener",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", ""),
			},
			expected: []v1.Service{
				service("finance", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": ""},
					nil),
			},
		},
		{
			name:        "named lb, no listener",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks"),
			},
			expected: []v1.Service{
				service("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					nil),
			},
		},
		{
			name:        "lb name has uppercase letters, no listener",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "STOCKS"),
			},
			expected: []v1.Service{
				service("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "STOCKS-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "STOCKS"},
					nil),
			},
		},
		{
			name:        "long lb name, no listener",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "H8O5dRuRZlLfzjOC4a6spBTnsNUmGGlVNCerKkeeK4w5qjgaDVa9ogLKBPJX539p"),
			},
			expected: []v1.Service{
				service("finance", "h8o5drurzllfzjoc4a6spbtn-9f88a4-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "H8O5dRuRZlLfzjOC4a6spBTnsNUmGGlVNCerKkeeK4w5qjgaDVa9ogLKBPJX539p-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "H8O5dRuRZlLfzjOC4a6spBTnsNUmGGlVNCerKkeeK4w5qjgaDVa9ogLKBPJX539p"},
					nil),
			},
		},
		{
			name:        "named lb, one listener",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("ls-1", "http", "tcp", "pool-1", 80)),
			},
			expected: []v1.Service{
				service("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.ServicePort{
						{
							Name:     "http-80",
							Port:     80,
							Protocol: v1.ProtocolTCP,
						},
					}),
			},
		},
		{
			name:        "named lb, one listener with uppercase name",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("ls-1", "HTTP", "tcp", "pool-1", 80)),
			},
			expected: []v1.Service{
				service("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.ServicePort{
						{
							Name:     "http-80",
							Port:     80,
							Protocol: v1.ProtocolTCP,
						},
					}),
			},
		},
		{
			name:        "named lb, one listener with long name",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("ls-1", "H8O5dRuRZlLfzjOC4a6spBTnsNUmGGlVNCerKkeeK4w5qjgaDVa9ogLKBPJX539p", "tcp", "pool-1", 80)),
			},
			expected: []v1.Service{
				service("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.ServicePort{
						{
							Name:     "h8o5drurzllfzjoc4a6spbtn-9c4814-80",
							Port:     80,
							Protocol: v1.ProtocolTCP,
						},
					}),
			},
		},
		{
			name:        "named lb, two listeners",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("ls-1", "http", "tcp", "pool-1", 80), listener("ls-1", "https", "tcp", "pool-1", 443)),
			},
			expected: []v1.Service{
				service("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.ServicePort{
						{
							Name:     "http-80",
							Port:     80,
							Protocol: v1.ProtocolTCP,
						},
						{
							Name:     "https-443",
							Port:     443,
							Protocol: v1.ProtocolTCP,
						},
					}),
			},
		},
		{
			name:        "unnammed lb, two listeners",
			tenantName:  "finance",
			clusterName: "us-east",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "", listener("ls-1", "http", "tcp", "pool-1", 80), listener("ls-1", "https", "tcp", "pool-1", 443)),
			},
			expected: []v1.Service{
				service("finance", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": ""},
					[]v1.ServicePort{
						{
							Name:     "http-80",
							Port:     80,
							Protocol: v1.ProtocolTCP,
						},
						{
							Name:     "https-443",
							Port:     443,
							Protocol: v1.ProtocolTCP,
						},
					}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := kubeServices(tc.clusterName, tc.tenantName, tc.lbs)
			assert.Equal(t, tc.expected, got)
			assert.Len(t, got, len(tc.lbs))
		})
	}
}

func TestKubeEndpoints(t *testing.T) {
	tests := []struct {
		name        string
		tenantName  string
		clusterName string
		lbs         []loadbalancers.LoadBalancer
		pools       []pools.Pool
		expected    []v1.Endpoints
	}{
		{
			name:        "named load balancer, no listener",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks"),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e", poolmember("10.0.0.1", 8080)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					nil),
			},
		},
		{
			name:        "long load balancer name, no listener",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "H8O5dRuRZlLfzjOC4a6spBTnsNUmGGlVNCerKkeeK4w5qjgaDVa9ogLKBPJX539p"),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e", poolmember("10.0.0.1", 8080)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "h8o5drurzllfzjoc4a6spbtn-9f88a4-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "H8O5dRuRZlLfzjOC4a6spBTnsNUmGGlVNCerKkeeK4w5qjgaDVa9ogLKBPJX539p-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "H8O5dRuRZlLfzjOC4a6spBTnsNUmGGlVNCerKkeeK4w5qjgaDVa9ogLKBPJX539p"},
					nil),
			},
		},
		{
			name:        "uppercase load balancer name, no listener",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "STOCKS"),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e", poolmember("10.0.0.1", 8080)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "STOCKS-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "STOCKS"},
					nil),
			},
		},
		{
			name:        "single listener",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("970fd223-4684-4c50-bfa2-738d6dda096f", "http", "tcp", "pool-1", 80)),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e", poolmember("10.0.0.1", 8080)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}},
							Ports:     []v1.EndpointPort{{Name: "http-80", Port: 8080, Protocol: v1.ProtocolTCP}},
						},
					}),
			},
		},
		{
			name:        "single listener with long name",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("970fd223-4684-4c50-bfa2-738d6dda096f", "H8O5dRuRZlLfzjOC4a6spBTnsNUmGGlVNCerKkeeK4w5qjgaDVa9ogLKBPJX539p", "tcp", "pool-1", 80)),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e", poolmember("10.0.0.1", 8080)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}},
							Ports:     []v1.EndpointPort{{Name: "h8o5drurzllfzjoc4a6spbtn-9c4814-80", Port: 8080, Protocol: v1.ProtocolTCP}},
						},
					}),
			},
		},
		{
			name:        "single listener with uppercase name",
			clusterName: "us-east",
			tenantName:  "finance",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("5a5c3d9e-e679-43ec-b9fc-9bc51132541e", "stocks", listener("970fd223-4684-4c50-bfa2-738d6dda096f", "HTTP", "tcp", "pool-1", 80)),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "5a5c3d9e-e679-43ec-b9fc-9bc51132541e", poolmember("10.0.0.1", 8080)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-id":   "5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}},
							Ports:     []v1.EndpointPort{{Name: "http-80", Port: 8080, Protocol: v1.ProtocolTCP}},
						},
					}),
			},
		},
		{
			name:        "multiple listeners",
			tenantName:  "finance",
			clusterName: "us-east",
			lbs: []loadbalancers.LoadBalancer{
				loadbalancer("loadbalancer-1", "stocks", listener("listener-1", "http", "tcp", "pool-1", 80), listener("listener-2", "https", "tcp", "pool-2", 443)),
			},
			pools: []pools.Pool{
				pool("pool-1", "HTTP", "loadbalancer-1", poolmember("10.0.0.1", 8080), poolmember("10.0.0.2", 80), poolmember("10.0.0.3", 80), poolmember("10.0.0.4", 8080)),
				pool("pool-2", "HTTP", "loadbalancer-1", poolmember("10.0.0.5", 443), poolmember("10.0.0.6", 443), poolmember("10.0.0.7", 8443)),
			},
			expected: []v1.Endpoints{
				endpoints("finance", "stocks-loadbalancer-1-us-east",
					map[string]string{
						"gimbal.heptio.com/cluster":            "us-east",
						"gimbal.heptio.com/service":            "stocks-loadbalancer-1",
						"gimbal.heptio.com/load-balancer-id":   "loadbalancer-1",
						"gimbal.heptio.com/load-balancer-name": "stocks"},
					[]v1.EndpointSubset{
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}, {IP: "10.0.0.4"}},
							Ports:     []v1.EndpointPort{{Name: "http-80", Port: 8080, Protocol: v1.ProtocolTCP}},
						},
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.2"}, {IP: "10.0.0.3"}},
							Ports:     []v1.EndpointPort{{Name: "http-80", Port: 80, Protocol: v1.ProtocolTCP}},
						},
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.5"}, {IP: "10.0.0.6"}},
							Ports:     []v1.EndpointPort{{Name: "https-443", Port: 443, Protocol: v1.ProtocolTCP}},
						},
						{
							Addresses: []v1.EndpointAddress{{IP: "10.0.0.7"}},
							Ports:     []v1.EndpointPort{{Name: "https-443", Port: 8443, Protocol: v1.ProtocolTCP}},
						},
					}),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := kubeEndpoints(tc.clusterName, tc.tenantName, tc.lbs, tc.pools)
			// Cannot use assert.Equal on the structs as the order of subsets is undetermined.
			for i := range tc.expected {
				assert.Equal(t, tc.expected[i].Namespace, got[i].Namespace)
				assert.Equal(t, tc.expected[i].Name, got[i].Name)
				assert.Equal(t, tc.expected[i].Labels, got[i].Labels)
				assert.ElementsMatch(t, tc.expected[i].Subsets, got[i].Subsets)
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
