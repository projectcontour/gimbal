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

	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/sirupsen/logrus"
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

func (action serviceAction) GetActionType() string {
	return action.kind
}

// Sync performs the action on the given service
func (action serviceAction) Sync(kubeClient kubernetes.Interface, logger *logrus.Logger) error {

	var err error
	switch action.kind {
	case actionAdd:
		err = addService(kubeClient, action.service)
	case actionUpdate:
		err = updateService(kubeClient, action.service)
	case actionDelete:
		err = deleteService(kubeClient, action.service)
	}
	if err != nil {
		return fmt.Errorf("error handling %s: %v", action, err)
	}

	return nil
}

func (action serviceAction) String() string {
	return fmt.Sprintf(`%s service '%s/%s'`, action.kind, action.service.Namespace, action.service.Name)
}

func addService(kubeClient kubernetes.Interface, service *v1.Service) error {
	_, err := kubeClient.CoreV1().Services(service.Namespace).Create(service)
	if errors.IsAlreadyExists(err) {
		return updateService(kubeClient, service)
	}
	return err
}

func deleteService(kubeClient kubernetes.Interface, service *v1.Service) error {
	return kubeClient.CoreV1().Services(service.Namespace).Delete(service.Name, &metav1.DeleteOptions{})
}

func updateService(kubeClient kubernetes.Interface, service *v1.Service) error {
	client := kubeClient.CoreV1().Services(service.Namespace)
	existing, err := client.Get(service.Name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			return addService(kubeClient, service)
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
	return err
}

func (action serviceAction) SetMetrics(gimbalKubeClient kubernetes.Interface, metrics localmetrics.DiscovererMetrics,
	logger *logrus.Logger) {

	// Log Service Event Timestamp
	metrics.ServiceEventTimestampMetric(action.service.GetNamespace(), action.service.GetName(), now().Unix())

	// Log Total Services Metric
	totalServices, err := getTotalServicesCount(gimbalKubeClient, action.ObjectMeta().GetNamespace(), metrics)
	if err != nil {
		logger.Error("Error getting total services count: ", err)
	} else {
		metrics.DiscovererReplicatedServicesMetric(action.service.GetNamespace(), totalServices)
	}
}

func (action serviceAction) SetMetricError(metrics localmetrics.DiscovererMetrics) {
	metrics.ServiceMetricError(action.ObjectMeta().GetNamespace(), action.ObjectMeta().GetName(), action.GetActionType())
}

// GetTotalServicesCount returns the number of services in a namespace for the particular backend
func getTotalServicesCount(gimbalKubeClient kubernetes.Interface, namespace string, metrics localmetrics.DiscovererMetrics) (int, error) {
	svcs, err := gimbalKubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("gimbal.heptio.com/backend=%s", metrics.BackendName)})
	if err != nil {
		return 0, err
	}
	return len(svcs.Items), nil
}
