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

package translator

import (
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
)

var quickCheckConfig = &quick.Config{MaxCount: 1000000}

func TestBuildDiscoveredName(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		cluster     string
		expected    string
	}{
		{
			name:        "basic",
			serviceName: "service1",
			cluster:     "cluster1",
			expected:    "cluster1-service1",
		},
		{
			name:        "long service name",
			serviceName: "the-really-long-kube-service-name-that-is-exactly-63-characters",
			cluster:     "cluster1",
			expected:    "cluster1-the-really-long-kube-serv1feeec",
		},
		{
			name:        "long cluster name",
			serviceName: "service1",
			cluster:     "a-really-long-cluster-name-that-does-not-really-make-sense-and-is-not-useful-at-all",
			expected:    "a-really-long-cluster-namfb8867-service1",
		},
		{
			name:        "long service and cluster names",
			serviceName: "the-really-long-kube-service-name-that-is-exactly-63-characters",
			cluster:     "a-really-long-cluster-name-that-does-not-really-make-sense-and-is-not-useful-at-all",
			expected:    "a-really-long-cluster-namfb8867-the-really-long-kube-serv1feeec",
		},
		{
			name:        "exact lengths, no shortening",
			serviceName: "name-that-is-exactly-at-d-limit",
			cluster:     "name-that-is-exactly-at-d-limit",
			expected:    "name-that-is-exactly-at-d-limit-name-that-is-exactly-at-d-limit",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := BuildDiscoveredName(test.cluster, test.serviceName)
			assert.Equal(t, test.expected, result, "Expected name does not match")
		})
	}
}

func TestBuildDiscoveredNameQuickTest(t *testing.T) {
	f := func(c, s string) bool {
		r := BuildDiscoveredName(c, s)
		// ensure discovered name does not go over limit
		if len(r) > maxKubernetesDNSLabelLength {
			return false
		}
		// ensure that no shortening occurs if the cluster name and service name are shorter than the limit
		// skip if the randomly generated string contained "-"
		maxPerComponent := (maxKubernetesDNSLabelLength - 1) / 2
		if len(c) <= maxPerComponent && len(s) <= maxPerComponent && strings.Count(r, "-") == 1 {
			comps := strings.Split(r, "-")
			return comps[0] == c && comps[1] == s
		}
		return true
	}
	if err := quick.Check(f, quickCheckConfig); err != nil {
		t.Error(err)
	}
}

func TestShortenKubernetesLabelValueQuickTest(t *testing.T) {
	f := func(s string) bool {
		r := ShortenKubernetesLabelValue(s)
		// ensure string size does not go over limit
		if len(r) > maxKubernetesDNSLabelLength {
			return false
		}
		// ensure that strings are not shortened if they are under the limit
		if len(s) <= maxKubernetesDNSLabelLength && r != s {
			return false
		}
		return true
	}
	if err := quick.Check(f, quickCheckConfig); err != nil {
		t.Error(err)
	}
}

func TestAddLabels(t *testing.T) {
	tests := []struct {
		name        string
		podLabels   map[string]string
		expected    map[string]string
		backendName string
		serviceName string
	}{
		{
			name:        "no existing labels",
			backendName: "test01",
			serviceName: "service01",
			podLabels:   nil,
			expected: map[string]string{
				"gimbal.heptio.com/backend": "test01",
				"gimbal.heptio.com/service": "service01",
			},
		},
		{
			name:        "simple test",
			backendName: "test01",
			serviceName: "service01",
			podLabels: map[string]string{
				"key1": "value1",
			},
			expected: map[string]string{
				"gimbal.heptio.com/backend": "test01",
				"gimbal.heptio.com/service": "service01",
				"key1": "value1",
			},
		},
		{
			name:        "heptio labels",
			backendName: "test01",
			serviceName: "service01",
			podLabels: map[string]string{
				"gimbal.heptio.com/backend": "badBackendName",
				"gimbal.heptio.com/service": "badService",
				"key1": "value1",
			},
			expected: map[string]string{
				"gimbal.heptio.com/backend": "test01",
				"gimbal.heptio.com/service": "service01",
				"key1": "value1",
			},
		},
		{
			name:        "long cluster name",
			backendName: "a-really-long-cluster-name-that-does-not-really-make-sense-and-is-not-useful-at-all",
			serviceName: "service01",
			podLabels:   nil,
			expected: map[string]string{
				"gimbal.heptio.com/backend": "a-really-long-cluster-name-that-does-not-really-make-sensfb8867",
				"gimbal.heptio.com/service": "service01",
			},
		},
		{
			name:        "long service name",
			backendName: "cluster01",
			serviceName: "a-really-long-service-name-that-does-not-really-make-sense-and-is-not-useful-at-all",
			podLabels:   nil,
			expected: map[string]string{
				"gimbal.heptio.com/backend": "cluster01",
				"gimbal.heptio.com/service": "a-really-long-service-name-that-does-not-really-make-sens1c0b9b",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AddGimbalLabels(test.backendName, test.serviceName, test.podLabels)
			assert.Equal(t, test.expected, result, "Expected name does not match")
		})
	}
}
