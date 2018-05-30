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

package sync

import (
	"fmt"
	"testing"
	"time"

	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestEndpointsAction(t *testing.T) {
	tests := []struct {
		name              string
		actionKind        string
		expectedVerbs     []string
		endpoints         v1.Endpoints
		existingEndpoints v1.Endpoints
		expectErr         bool
	}{
		{
			name:          "add new endpoints resource",
			actionKind:    actionAdd,
			endpoints:     v1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs: []string{"create"},
		},
		{
			name:              "add pre-existing endpoints resource",
			actionKind:        actionAdd,
			endpoints:         v1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			existingEndpoints: v1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs:     []string{"create", "get", "patch"},
		},
		{
			name:              "update pre-existing endpoints resource",
			actionKind:        actionUpdate,
			endpoints:         v1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			existingEndpoints: v1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs:     []string{"get", "patch"},
		},
		{
			name:          "update non-existent endpoints resource",
			actionKind:    actionUpdate,
			endpoints:     v1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectErr:     false,
			expectedVerbs: []string{"get", "create"},
		},
		{
			name:              "delete endpoints resource",
			actionKind:        actionDelete,
			endpoints:         v1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			existingEndpoints: v1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs:     []string{"delete"},
		},
		{
			name:          "delete non-existent endpoints resource",
			actionKind:    actionDelete,
			endpoints:     v1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs: []string{"delete"},
			expectErr:     true,
		},
	}

	expectedResource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "endpoints"}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewSimpleClientset(&tc.existingEndpoints)
			a := endpointsAction{kind: tc.actionKind, endpoints: &tc.endpoints}
			err := a.Sync(client, logrus.New())

			if !tc.expectErr {
				require.NoError(t, err)
			}
			require.Len(t, client.Actions(), len(tc.expectedVerbs))
			for i, expectedVerb := range tc.expectedVerbs {
				assert.Equal(t, expectedResource, client.Actions()[i].GetResource())
				assert.Equal(t, expectedVerb, client.Actions()[i].GetVerb())
			}
		})
	}
}

func TestUpdateEndpoints(t *testing.T) {
	existing := v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foo",
			Name:      "bar",
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{{IP: "192.168.0.1"}},
				Ports:     []v1.EndpointPort{{Port: 80}},
			},
		},
	}

	var gotPatchBytes []byte
	client := fake.NewSimpleClientset(&existing)
	client.PrependReactor("patch", "endpoints", func(action k8stesting.Action) (bool, runtime.Object, error) {
		switch patchAction := action.(type) {
		default:
			return true, nil, fmt.Errorf("got unexpected action of type: %T", action)
		case k8stesting.PatchActionImpl:
			gotPatchBytes = patchAction.GetPatch()
			return true, &existing, nil
		}
	})

	newEndpoints := existing
	newEndpoints.Subsets = []v1.EndpointSubset{
		{
			Addresses: []v1.EndpointAddress{{IP: "192.168.0.2"}},
			Ports:     []v1.EndpointPort{{Port: 8080}},
		},
		{
			Addresses: []v1.EndpointAddress{{IP: "192.168.0.3"}},
			Ports:     []v1.EndpointPort{{Port: 80}},
		},
	}
	expectedPatch := `{"subsets":[{"addresses":[{"ip":"192.168.0.2"}],"ports":[{"port":8080}]},{"addresses":[{"ip":"192.168.0.3"}],"ports":[{"port":80}]}]}`
	err := updateEndpoints(client, &newEndpoints)
	require.NoError(t, err)
	assert.Equal(t, expectedPatch, string(gotPatchBytes))
}

func TestDiscovererEndpointsMetrics(t *testing.T) {
	backendName := "backend"
	backendType := "backtype"
	tests := []struct {
		name              string
		actionKind        string
		endpoints         v1.Endpoints
		existingendpoints v1.Endpoints
		expectErr         bool
		expectedCount     float64
		expectedTimestamp float64
	}{
		{
			name:       "add new endpoint resource",
			actionKind: actionAdd,
			endpoints: v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
					Labels: map[string]string{
						"gimbal.heptio.com/backend": backendName,
					},
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{{IP: "192.168.0.1"}},
						Ports:     []v1.EndpointPort{{Port: 80}},
					},
				},
			},
			existingendpoints: v1.Endpoints{},
			expectedCount:     float64(1),
			expectedTimestamp: 9.467208e+08,
		},
		{
			name:       "add new endpoint resource, multiple ips",
			actionKind: actionAdd,
			endpoints: v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
					Labels: map[string]string{
						"gimbal.heptio.com/backend": backendName,
					},
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{{IP: "192.168.0.1"}},
						Ports:     []v1.EndpointPort{{Port: 80}},
					},
					{
						Addresses: []v1.EndpointAddress{{IP: "192.168.0.2"}, {IP: "192.168.0.3"}},
						Ports:     []v1.EndpointPort{{Port: 443}},
					},
				},
			},
			existingendpoints: v1.Endpoints{},
			expectedCount:     float64(3),
			expectedTimestamp: 9.467208e+08,
		},
		{
			name:       "add new endpoint resource, existing-non gimbal",
			actionKind: actionAdd,
			endpoints: v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
					Labels: map[string]string{
						"gimbal.heptio.com/backend": backendName,
					},
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{{IP: "192.168.0.1"}, {IP: "192.168.0.2"}},
						Ports:     []v1.EndpointPort{{Port: 80}},
					},
				},
			},
			existingendpoints: v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "ep-existing",
					Labels: map[string]string{
						"who": "dis",
					},
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{{IP: "192.168.0.1"}},
						Ports:     []v1.EndpointPort{{Port: 80}},
					},
				},
			},
			expectedCount:     float64(2),
			expectedTimestamp: 9.467208e+08,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nowFunc = func() time.Time {
				return time.Date(2000, 1, 1, 10, 0, 00, 0, time.UTC)
			}
			client := fake.NewSimpleClientset(&tc.existingendpoints)
			metrics := localmetrics.NewMetrics(backendType, backendName)
			metrics.RegisterPrometheus(false)
			a := endpointsAction{kind: tc.actionKind, endpoints: &tc.endpoints}

			err := a.Sync(client, logrus.New())

			if !tc.expectErr {
				require.NoError(t, err)
			}

			a.SetMetrics(client, metrics, logrus.New())

			gatherers := prometheus.Gatherers{
				metrics.Registry,
				prometheus.DefaultGatherer,
			}

			gathering, err := gatherers.Gather()
			if err != nil {
				t.Fatal(err)
			}

			replicatedEndpoints := float64(-1)
			timestamp := float64(-1)
			for _, mf := range gathering {
				if mf.GetName() == localmetrics.DiscovererReplicatedEndpointsGauge {
					replicatedEndpoints = mf.Metric[0].Gauge.GetValue()
				} else if mf.GetName() == localmetrics.EndpointsEventTimestampGauge {
					timestamp = mf.Metric[0].Gauge.GetValue()
				}
			}

			assert.Equal(t, tc.expectedCount, replicatedEndpoints)
			assert.Equal(t, tc.expectedTimestamp, timestamp)
		})
	}

}
