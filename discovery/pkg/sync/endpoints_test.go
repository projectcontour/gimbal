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

	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			err := a.Sync(client, localmetrics.NewMetrics())

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
	err := updateEndpoints(client, &newEndpoints, localmetrics.NewMetrics())
	require.NoError(t, err)
	assert.Equal(t, expectedPatch, string(gotPatchBytes))
}
