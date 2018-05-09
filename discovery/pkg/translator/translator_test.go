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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildKubernetesDNSLabel(t *testing.T) {
	tests := []struct {
		name       string
		components []string
		expected   string
	}{
		{
			name:       "basic",
			components: []string{"service1", "cluster1"},
			expected:   "service1-cluster1",
		},
		{
			name: "one long name",
			components: []string{
				"service1service1service1service1service1service1service1service1service1service1service1service1service1",
				"cluster1",
			},
			expected: "service1service1service1-d8cb7f-cluster1",
		},
		{
			name: "both long names",
			components: []string{
				"service1service1service1service1service1service1service1service1service1service1service1service1service1",
				"cluster1cluster1cluster1cluster1cluster1cluster1cluster1cluster1",
			},
			expected: "4ce97d89bfa193a277ddd97df3cad58484b30bb4bc4b815ea71c84518d9a830",
		},
		{
			name:       "uppercase characters in one of the components",
			components: []string{"FOO", "bar"},
			expected:   "foo-bar",
		},
		{
			name:       "one uppercase component",
			components: []string{"FOO"},
			expected:   "foo",
		},
		{
			name:       "one long component",
			components: []string{"service1service1service1service1service1service1service1service1service1service1service1service1service1"},
			expected:   "bbaff9d26694ea6a881ebd3daf025e0daf0921f7398789053aeef0ac39be935",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := BuildKubernetesDNSLabel(test.components...)

			assert.Equal(t, test.expected, result, "Expected name does not match")
		})
	}
}

func TestAddLabels(t *testing.T) {
	clusterName := "test01"
	namespace := "default"
	serviceName := "service01"

	tests := []struct {
		name      string
		podLabels map[string]string
		expected  map[string]string
	}{
		{
			name:      "no existing labels",
			podLabels: nil,
			expected: map[string]string{
				"gimbal.heptio.com/cluster": clusterName,
				"gimbal.heptio.com/service": serviceName,
			},
		},
		{
			name: "simple test",
			podLabels: map[string]string{
				"key1": "value1",
			},
			expected: map[string]string{
				"gimbal.heptio.com/cluster": clusterName,
				"gimbal.heptio.com/service": serviceName,
				"key1": "value1",
			},
		},
		{
			name: "heptio labels",
			podLabels: map[string]string{
				"gimbal.heptio.com/cluster": "badClusterName",
				"gimbal.heptio.com/service": "badService",
				"key1": "value1",
			},
			expected: map[string]string{
				"gimbal.heptio.com/cluster": clusterName,
				"gimbal.heptio.com/service": serviceName,
				"key1": "value1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AddGimbalLabels(clusterName, namespace, serviceName, test.podLabels)

			assert.Equal(t, test.expected, result, "Expected name does not match")
		})
	}
}
