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
	"encoding/json"
	"fmt"
	"time"

	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
)

// AddEndpointsAction returns an action that adds a new endpoint to the cluster
func AddEndpointsAction(endpoints *v1.Endpoints) Action {
	return endpointsAction{kind: actionAdd, endpoints: endpoints}
}

// UpdateEndpointsAction returns an action that updates the given endpoint in the cluster
func UpdateEndpointsAction(endpoints *v1.Endpoints) Action {
	return endpointsAction{kind: actionUpdate, endpoints: endpoints}
}

// DeleteEndpointsAction returns an action that deletes the given endpoint from the cluster
func DeleteEndpointsAction(endpoints *v1.Endpoints) Action {
	return endpointsAction{kind: actionDelete, endpoints: endpoints}
}

// endpointsAction is an action that is to be performed on a specific endpoint.
type endpointsAction struct {
	kind      string
	endpoints *v1.Endpoints
}

// ObjectMeta returns the objectMeta piece of the Action interface object
func (action endpointsAction) ObjectMeta() *metav1.ObjectMeta {
	return &action.endpoints.ObjectMeta
}

// Sync performs the action on the given Endpoints resource
func (action endpointsAction) Sync(kubeClient kubernetes.Interface, metrics localmetrics.DiscovererMetrics, clusterName string) error {

	var err error
	switch action.kind {
	case actionAdd:
		err = addEndpoints(kubeClient, action.endpoints, metrics, clusterName)
	case actionUpdate:
		err = updateEndpoints(kubeClient, action.endpoints, metrics, clusterName)
	case actionDelete:
		err = deleteEndpoints(kubeClient, action.endpoints, metrics, clusterName)
	}
	if err != nil {
		return fmt.Errorf("error handling %s: %v", action, err)
	}

	metrics.EndpointsEventTimestampMetric(action.endpoints.GetNamespace(), clusterName, action.endpoints.GetName(), time.Now().Unix())
	return nil
}

func (action endpointsAction) String() string {
	return fmt.Sprintf(`%s endpoints "%s/%s"`, action.kind, action.endpoints.Namespace, action.endpoints.Name)
}

func addEndpoints(kubeClient kubernetes.Interface, endpoints *v1.Endpoints, lm localmetrics.DiscovererMetrics, clusterName string) error {
	_, err := kubeClient.CoreV1().Endpoints(endpoints.Namespace).Create(endpoints)
	if errors.IsAlreadyExists(err) {
		err = updateEndpoints(kubeClient, endpoints, lm, clusterName)
		if err != nil {
			lm.EndpointsMetricError(endpoints.GetNamespace(), clusterName, endpoints.GetName(), "UPDATE")
		}
	} else {
		if err != nil {
			lm.EndpointsMetricError(endpoints.GetNamespace(), clusterName, endpoints.GetName(), "ADD")
		}
	}
	return err
}

func deleteEndpoints(kubeClient kubernetes.Interface, endpoints *v1.Endpoints, lm localmetrics.DiscovererMetrics, clusterName string) error {
	err := kubeClient.CoreV1().Endpoints(endpoints.Namespace).Delete(endpoints.Name, &metav1.DeleteOptions{})

	if err != nil {
		lm.EndpointsMetricError(endpoints.GetNamespace(), clusterName, endpoints.GetName(), "DELETE")
	}

	return err
}

func updateEndpoints(kubeClient kubernetes.Interface, endpoints *v1.Endpoints, lm localmetrics.DiscovererMetrics, clusterName string) error {
	client := kubeClient.CoreV1().Endpoints(endpoints.Namespace)
	existing, err := client.Get(endpoints.Name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			err = addEndpoints(kubeClient, endpoints, lm, clusterName)
			if err != nil {
				lm.EndpointsMetricError(endpoints.GetNamespace(), clusterName, endpoints.GetName(), "ADD")
			}
		}
		return err
	}

	existingBytes, err := json.Marshal(existing)
	if err != nil {
		return err
	}
	// Need to set the resource version of the updated endpoints to the resource
	// version of the current service. Otherwise, the resulting patch does not
	// have a resource version, and the server complains.
	endpoints.ResourceVersion = existing.ResourceVersion
	updatedBytes, err := json.Marshal(endpoints)
	if err != nil {
		return err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(existingBytes, updatedBytes, v1.Endpoints{})
	if err != nil {
		return err
	}
	_, err = client.Patch(endpoints.Name, types.MergePatchType, patchBytes)

	if err != nil {
		lm.EndpointsMetricError(endpoints.GetNamespace(), clusterName, endpoints.GetName(), "UPDATE")
	}

	return err
}
