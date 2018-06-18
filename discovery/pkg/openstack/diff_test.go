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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDiffServices(t *testing.T) {
	tests := []struct {
		name           string
		current        []v1.Service
		desired        []v1.Service
		expectedAdd    []v1.Service
		expectedUpdate []v1.Service
		expectedDel    []v1.Service
	}{
		{
			name: "new service",
			desired: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "production-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					},
				},
			},
			expectedAdd: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "production-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
					},
				},
			},
		},
		{
			name: "updated service",
			current: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "production",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "http",
								Port:     80,
								Protocol: "TCP",
							},
						},
					},
				},
			},
			desired: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "production",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "https",
								Port:     443,
								Protocol: "TCP",
							},
						},
					},
				},
			},
			expectedUpdate: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "production",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "https",
								Port:     443,
								Protocol: "TCP",
							},
						},
					},
				},
			},
		},
		{
			name: "deleted service",
			current: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "production",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "http",
								Port:     80,
								Protocol: "TCP",
							},
						},
					},
				},
			},
			expectedDel: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "production",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "http",
								Port:     80,
								Protocol: "TCP",
							},
						},
					},
				},
			},
		},
		{
			name: "order doesn't matter for update",
			current: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "service1",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "http",
								Port:     80,
								Protocol: "TCP",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "service2",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "http",
								Port:     80,
								Protocol: "TCP",
							},
						},
					},
				},
			},
			desired: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "service2",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "http",
								Port:     80,
								Protocol: "TCP",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "finance",
						Name:      "service1",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "http",
								Port:     80,
								Protocol: "TCP",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			add, up, del := diffServices(tc.desired, tc.current)
			assert.Equal(t, tc.expectedAdd, add, "ExpectedADD")
			assert.Equal(t, tc.expectedUpdate, up, "ExpectedUPDATE")
			assert.Equal(t, tc.expectedDel, del, "ExpectedDELETE")
		})
	}
}
func TestDiffEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		current        []Endpoints
		desired        []Endpoints
		expectedAdd    []Endpoints
		expectedUpdate []Endpoints
		expectedDel    []Endpoints
	}{
		{
			name: "new endpoint",
			desired: []Endpoints{
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "production-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						},
					},
					upstreamName: "upname",
				},
			},
			expectedAdd: []Endpoints{
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "production-stocks-5a5c3d9e-e679-43ec-b9fc-9bc51132541e",
						},
					},
					upstreamName: "upname",
				},
			},
		},
		{
			name: "updated endpoint",
			current: []Endpoints{
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "production",
						},
						Subsets: []v1.EndpointSubset{
							{
								Addresses: []v1.EndpointAddress{
									{
										IP: "5.6.7.8",
									},
								},
								Ports: []v1.EndpointPort{
									{
										Name:     "svc2",
										Port:     443,
										Protocol: v1.ProtocolTCP,
									},
								},
							},
						},
					},
					upstreamName: "upname",
				},
			},
			desired: []Endpoints{
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "production",
						},
						Subsets: []v1.EndpointSubset{
							{
								Addresses: []v1.EndpointAddress{
									{
										IP: "1.2.3.4",
									},
								},
								Ports: []v1.EndpointPort{
									{
										Name:     "svc1",
										Port:     80,
										Protocol: v1.ProtocolTCP,
									},
								},
							},
						},
					},
					upstreamName: "upname",
				},
			},
			expectedUpdate: []Endpoints{
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "production",
						},
						Subsets: []v1.EndpointSubset{
							{
								Addresses: []v1.EndpointAddress{
									{
										IP: "1.2.3.4",
									},
								},
								Ports: []v1.EndpointPort{
									{
										Name:     "svc1",
										Port:     80,
										Protocol: v1.ProtocolTCP,
									},
								},
							},
						},
					},
					upstreamName: "upname",
				},
			},
		},
		{
			name: "deleted service",
			current: []Endpoints{
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "production",
						},
						Subsets: []v1.EndpointSubset{
							{
								Addresses: []v1.EndpointAddress{
									{
										IP: "1.2.3.4",
									},
								},
								Ports: []v1.EndpointPort{
									{
										Name:     "svc1",
										Port:     80,
										Protocol: v1.ProtocolTCP,
									},
								},
							},
						},
					},
					upstreamName: "upname",
				},
			},
			expectedDel: []Endpoints{
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "production",
						},
						Subsets: []v1.EndpointSubset{
							{
								Addresses: []v1.EndpointAddress{
									{
										IP: "1.2.3.4",
									},
								},
								Ports: []v1.EndpointPort{
									{
										Name:     "svc1",
										Port:     80,
										Protocol: v1.ProtocolTCP,
									},
								},
							},
						},
					},
					upstreamName: "upname",
				},
			},
		},
		{
			name: "order doesn't matter for update",
			current: []Endpoints{
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "endpoints1",
						},
						Subsets: []v1.EndpointSubset{
							{
								Addresses: []v1.EndpointAddress{
									{
										IP: "1.2.3.4",
									},
								},
								Ports: []v1.EndpointPort{
									{
										Name:     "svc1",
										Port:     80,
										Protocol: v1.ProtocolTCP,
									},
								},
							},
						},
					},
					upstreamName: "",
				},
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "endpoints2",
						},
						Subsets: []v1.EndpointSubset{
							{
								Addresses: []v1.EndpointAddress{
									{
										IP: "1.2.3.4",
									},
								},
								Ports: []v1.EndpointPort{
									{
										Name:     "svc1",
										Port:     80,
										Protocol: v1.ProtocolTCP,
									},
								},
							},
						},
					},
					upstreamName: "upname",
				},
			},
			desired: []Endpoints{
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "endpoints2",
						},
						Subsets: []v1.EndpointSubset{
							{
								Addresses: []v1.EndpointAddress{
									{
										IP: "1.2.3.4",
									},
								},
								Ports: []v1.EndpointPort{
									{
										Name:     "svc1",
										Port:     80,
										Protocol: v1.ProtocolTCP,
									},
								},
							},
						},
					},
					upstreamName: "upname",
				},
				{
					endpoints: v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "finance",
							Name:      "endpoints1",
						},
						Subsets: []v1.EndpointSubset{
							{
								Addresses: []v1.EndpointAddress{
									{
										IP: "1.2.3.4",
									},
								},
								Ports: []v1.EndpointPort{
									{
										Name:     "svc1",
										Port:     80,
										Protocol: v1.ProtocolTCP,
									},
								},
							},
						},
					},
					upstreamName: "upname",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			add, up, del := diffEndpoints(tc.desired, tc.current)
			assert.Equal(t, tc.expectedAdd, add, "ExpectedADD")
			assert.Equal(t, tc.expectedUpdate, up, "ExpectedUPDATE")
			assert.Equal(t, tc.expectedDel, del, "ExpectedDELETE")
		})
	}
}
