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

// AddServiceAction returns an action that adds a new endpoint to the cluster
func AddServiceAction(service *v1.Service) Action {
	return serviceAction{kind: actionAdd, service: service}
}

// UpdateServiceAction returns an action that updates the given endpoint in the cluster
func UpdateServiceAction(service *v1.Service) Action {
	return serviceAction{kind: actionUpdate, service: service}
}

// DeleteServiceAction returns an action that deletes the given endpoint from the cluster
func DeleteServiceAction(service *v1.Service) Action {
	return serviceAction{kind: actionDelete, service: service}
}

// serviceAction is an action that should be performed on a given service
type serviceAction struct {
	kind    string
	service *v1.Service
}

// ObjectMeta returns the objectMeta piece of the Action interface object
func (action serviceAction) ObjectMeta() *metav1.ObjectMeta {
	return &action.service.ObjectMeta
}

// Sync performs the action on the given service
func (action serviceAction) Sync(kubeClient kubernetes.Interface, metrics localmetrics.DiscovererMetrics, clusterName string) error {

	var err error
	switch action.kind {
	case actionAdd:
		err = addService(kubeClient, action.service, metrics, clusterName)
	case actionUpdate:
		err = updateService(kubeClient, action.service, metrics, clusterName)
	case actionDelete:
		err = deleteService(kubeClient, action.service, metrics, clusterName)
	}
	if err != nil {
		return fmt.Errorf("error handling %s: %v", action, err)
	}

	metrics.ServiceEventTimestampMetric(action.service.GetNamespace(), clusterName, action.service.GetName(), time.Now().Unix())
	return nil
}

func (action serviceAction) String() string {
	return fmt.Sprintf(`%s service "%s/%s"`, action.kind, action.service.Namespace, action.service.Name)
}

func addService(kubeClient kubernetes.Interface, service *v1.Service, lm localmetrics.DiscovererMetrics, clusterName string) error {
	_, err := kubeClient.CoreV1().Services(service.Namespace).Create(service)
	if errors.IsAlreadyExists(err) {
		err = updateService(kubeClient, service, lm, clusterName)
		if err != nil {
			lm.ServiceMetricError(service.GetNamespace(), clusterName, service.GetName(), "UPDATE")
		}
	} else {
		if err != nil {
			lm.ServiceMetricError(service.GetNamespace(), clusterName, service.GetName(), "ADD")
		}
	}
	return err
}

func deleteService(kubeClient kubernetes.Interface, service *v1.Service, lm localmetrics.DiscovererMetrics, clusterName string) error {
	err := kubeClient.CoreV1().Services(service.Namespace).Delete(service.Name, &metav1.DeleteOptions{})

	if err != nil {
		lm.ServiceMetricError(service.GetNamespace(), clusterName, service.GetName(), "DELETE")
	}
	return err
}

func updateService(kubeClient kubernetes.Interface, service *v1.Service, lm localmetrics.DiscovererMetrics, clusterName string) error {
	client := kubeClient.CoreV1().Services(service.Namespace)
	existing, err := client.Get(service.Name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			err = addService(kubeClient, service, lm, clusterName)
			if err != nil {
				lm.ServiceMetricError(service.GetNamespace(), clusterName, service.GetName(), "ADD")
			}
			return err
		}
		return err
	}

	existingBytes, err := json.Marshal(existing)
	if err != nil {
		return err
	}
	// Need to set the resource version of the updated service to the resource
	// version of the current service. Otherwise, the resulting patch does not
	// have a resource version, and the server complains.
	service.ResourceVersion = existing.ResourceVersion
	updatedBytes, err := json.Marshal(service)
	if err != nil {
		return err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(existingBytes, updatedBytes, v1.Service{})
	if err != nil {
		return err
	}
	_, err = client.Patch(service.Name, types.StrategicMergePatchType, patchBytes)

	if err != nil {
		lm.ServiceMetricError(service.GetNamespace(), clusterName, service.GetName(), "UPDATE")
	}

	return err
}
