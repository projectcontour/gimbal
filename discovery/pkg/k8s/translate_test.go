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

	"github.com/stretchr/testify/assert"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestTranslateService(t *testing.T) {
	tests := []struct {
		name        string
		backendName string
		service     *v1.Service
		expected    *v1.Service
	}{
		{
			name:        "simple service",
			backendName: "cluster1",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "kuard",
					Labels:    map[string]string{"app": "kuard"},
				},
				Spec: v1.ServiceSpec{
					ClusterIP: "10.99.179.252",
					Ports:     []v1.ServicePort{{Name: "foo", Port: 80, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)}},
					Selector:  map[string]string{"app": "kuard"},
					Type:      v1.ServiceTypeClusterIP,
				},
			},
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "cluster1-kuard",
					Labels:    map[string]string{"app": "kuard", "gimbal.heptio.com/backend": "cluster1", "gimbal.heptio.com/service": "kuard"},
				},
				Spec: v1.ServiceSpec{
					ClusterIP: "None",
					Ports:     []v1.ServicePort{{Name: "foo", Port: 80}}, //, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)}},
					Type:      v1.ServiceTypeClusterIP,
				},
			},
		},
		{
			name:        "multi-port service",
			backendName: "cluster1",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "kuard",
					Labels:    map[string]string{"app": "kuard"},
				},
				Spec: v1.ServiceSpec{
					ClusterIP: "10.99.179.252",
					Ports: []v1.ServicePort{
						{Name: "foo", Port: 80, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)},
						{Name: "bar", Port: 8080, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)},
					},
					Selector: map[string]string{"app": "kuard"},
					Type:     v1.ServiceTypeClusterIP,
				},
			},
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "cluster1-kuard",
					Labels:    map[string]string{"app": "kuard", "gimbal.heptio.com/backend": "cluster1", "gimbal.heptio.com/service": "kuard"},
				},
				Spec: v1.ServiceSpec{
					ClusterIP: "None",
					Ports: []v1.ServicePort{
						{Name: "foo", Port: 80},
						{Name: "bar", Port: 8080},
					},
					Type: v1.ServiceTypeClusterIP,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := translateService(tc.service, tc.backendName)
			assert.EqualValues(t, tc.expected, got)
		})
	}
}

func TestTranslateEndpoints(t *testing.T) {
	nodeName := "minikube"
	tests := []struct {
		name        string
		backendName string
		endpoints   *v1.Endpoints
		expected    *v1.Endpoints
	}{
		{
			name:        "simple endpoints",
			backendName: "cluster1",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "kuard",
					Labels:    map[string]string{"app": "kuard"},
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{{IP: "172.17.0.4", NodeName: &nodeName}, {IP: "172.17.0.7", NodeName: &nodeName}, {IP: "172.17.0.9", NodeName: &nodeName}},
						Ports:     []v1.EndpointPort{{Name: "foo", Port: 8080, Protocol: v1.ProtocolTCP}},
					},
				},
			},
			expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "cluster1-kuard",
					Labels:    map[string]string{"app": "kuard", "gimbal.heptio.com/backend": "cluster1", "gimbal.heptio.com/service": "kuard"},
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{{IP: "172.17.0.4", NodeName: &nodeName}, {IP: "172.17.0.7", NodeName: &nodeName}, {IP: "172.17.0.9", NodeName: &nodeName}},
						Ports:     []v1.EndpointPort{{Name: "foo", Port: 8080, Protocol: v1.ProtocolTCP}},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := translateEndpoints(tc.endpoints, tc.backendName)
			assert.EqualValues(t, tc.expected, got)
		})
	}
}
