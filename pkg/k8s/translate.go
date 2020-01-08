// Copyright © 2018 the Gimbal contributors.
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
	"github.com/projectcontour/gimbal/pkg/translator"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func translateService(svc *v1.Service, backendName string) *v1.Service {
	newService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   svc.Namespace,
			Name:        translator.BuildDiscoveredName(backendName, svc.Name),
			Labels:      translator.AddGimbalLabels(backendName, svc.ObjectMeta.Name, svc.ObjectMeta.Labels),
			Annotations: svc.Annotations,
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

func translateEndpoints(endpoints *v1.Endpoints, backendName string) *v1.Endpoints {
	newEndpoint := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   endpoints.Namespace,
			Name:        translator.BuildDiscoveredName(backendName, endpoints.Name),
			Labels:      translator.AddGimbalLabels(backendName, endpoints.ObjectMeta.Name, endpoints.ObjectMeta.Labels),
			Annotations: endpoints.Annotations,
		},
		Subsets: endpoints.Subsets,
	}
	return newEndpoint
}
