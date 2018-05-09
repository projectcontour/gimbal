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
	"github.com/heptio/gimbal/discovery/pkg/translator"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func translateService(svc *v1.Service, clusterName string) *v1.Service {
	newService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: svc.Namespace,
			Name:      translator.BuildKubernetesDNSLabel(svc.Name, clusterName),
			Labels:    translator.AddGimbalLabels(clusterName, svc.Namespace, svc.ObjectMeta.Name, svc.ObjectMeta.Labels),
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Type:      v1.ServiceTypeClusterIP,
		},
	}

	for _, port := range svc.Spec.Ports {
		newService.Spec.Ports = append(newService.Spec.Ports, v1.ServicePort{
			Name: port.Name,
			Port: port.Port,
		})
	}
	return newService
}

func translateEndpoints(endpoints *v1.Endpoints, clusterName string) *v1.Endpoints {
	newEndpoint := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: endpoints.Namespace,
			Name:      translator.BuildKubernetesDNSLabel(endpoints.Name, clusterName),
			Labels:    translator.AddGimbalLabels(clusterName, endpoints.Namespace, endpoints.ObjectMeta.Name, endpoints.ObjectMeta.Labels),
		},
		Subsets: endpoints.Subsets,
	}
	return newEndpoint
}
