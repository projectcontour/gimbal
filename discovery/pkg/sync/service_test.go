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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestServiceActions(t *testing.T) {
	tests := []struct {
		name            string
		actionKind      string
		expectedVerbs   []string
		service         v1.Service
		existingService v1.Service
		expectErr       bool
	}{
		{
			name:          "add new service resource",
			actionKind:    actionAdd,
			service:       v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs: []string{"create"},
		},
		{
			name:            "add pre-existing service resource",
			actionKind:      actionAdd,
			service:         v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			existingService: v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs:   []string{"create", "get", "patch"},
		},
		{
			name:            "update pre-existing service resource",
			actionKind:      actionUpdate,
			service:         v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			existingService: v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs:   []string{"get", "patch"},
		},
		{
			name:          "update non-existent service resource",
			actionKind:    actionUpdate,
			service:       v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectErr:     false,
			expectedVerbs: []string{"get", "create"},
		},
		{
			name:            "delete service resource",
			actionKind:      actionDelete,
			service:         v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			existingService: v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs:   []string{"delete"},
		},
		{
			name:          "delete non-existent service resource",
			actionKind:    actionDelete,
			service:       v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}},
			expectedVerbs: []string{"delete"},
			expectErr:     true,
		},
	}

	expectedResource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewSimpleClientset(&tc.existingService)
			a := serviceAction{kind: tc.actionKind, service: &tc.service}
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

func TestUpdateService(t *testing.T) {
	existing := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foo",
			Name:      "bar",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}

	var gotPatchBytes []byte
	client := fake.NewSimpleClientset(&existing)
	client.PrependReactor("patch", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		switch patchAction := action.(type) {
		default:
			return true, nil, fmt.Errorf("got unexpected action of type: %T", action)
		case k8stesting.PatchActionImpl:
			gotPatchBytes = patchAction.GetPatch()
			return true, &existing, nil
		}
	})

	newService := existing
	newService.Spec = v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Port: 8080,
			},
		},
	}
	expectedPatch := `{"spec":{"$setElementOrder/ports":[{"port":8080}],"ports":[{"port":8080,"targetPort":0},{"$patch":"delete","port":80}]}}`
	err := updateService(client, &newService)
	require.NoError(t, err)
	assert.Equal(t, expectedPatch, string(gotPatchBytes))
}

func TestDiscovererServiceMetrics(t *testing.T) {
	backendName := "backend"
	backendType := "backtype"
	tests := []struct {
		name              string
		actionKind        string
		service           v1.Service
		existingservice   v1.Service
		expectErr         bool
		expectedCount     float64
		expectedTimestamp float64
		expectedLabels    map[string]string
	}{
		{
			name:       "add new service resource",
			actionKind: actionAdd,
			service: v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
					Labels: map[string]string{
						"gimbal.heptio.com/backend": backendName,
					},
				},
			},
			existingservice:   v1.Service{},
			expectedCount:     float64(1),
			expectedTimestamp: 9.467208e+08,
			expectedLabels: map[string]string{
				"backendname": backendName,
				"backendtype": backendType,
				"namespace":   "foo",
			},
		},
		{
			name:       "add new service resource with existing",
			actionKind: actionAdd,
			service: v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
					Labels: map[string]string{
						"gimbal.heptio.com/backend": backendName,
					},
				},
			},
			existingservice: v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "existing",
					Labels: map[string]string{
						"gimbal.heptio.com/backend": backendName,
					},
				},
			},
			expectedCount:     float64(2),
			expectedTimestamp: 9.467208e+08,
			expectedLabels: map[string]string{
				"backendname": backendName,
				"backendtype": backendType,
				"namespace":   "foo",
			},
		},
		{
			name:       "add new service resource with existing non-gimbal",
			actionKind: actionAdd,
			service: v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
					Labels: map[string]string{
						"gimbal.heptio.com/backend": backendName,
					},
				},
			},
			existingservice: v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "existing",
					Labels: map[string]string{
						"key": "value",
					},
				},
			},
			expectedCount:     float64(1),
			expectedTimestamp: 9.467208e+08,
			expectedLabels: map[string]string{
				"backendname": backendName,
				"backendtype": backendType,
				"namespace":   "foo",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nowFunc = func() time.Time {
				return time.Date(2000, 1, 1, 10, 0, 00, 0, time.UTC)
			}
			client := fake.NewSimpleClientset(&tc.existingservice)
			metrics := localmetrics.NewMetrics(backendType, backendName)
			metrics.RegisterPrometheus(false)
			a := serviceAction{kind: tc.actionKind, service: &tc.service}

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
				fmt.Println(err)
			}

			replicatedServices := float64(-1)
			timestamp := float64(-1)
			for _, mf := range gathering {
				if mf.GetName() == localmetrics.DiscovererReplicatedServicesGauge {
					replicatedServices = mf.Metric[0].Gauge.GetValue()
				} else if mf.GetName() == localmetrics.ServiceEventTimestampGauge {
					timestamp = mf.Metric[0].Gauge.GetValue()
				}
			}

			assert.Equal(t, tc.expectedCount, replicatedServices)
			assert.Equal(t, tc.expectedTimestamp, timestamp)
		})
	}

}
