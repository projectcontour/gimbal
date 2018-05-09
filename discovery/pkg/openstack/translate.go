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

package openstack

import (
	"fmt"
	"strconv"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/pools"
	"github.com/heptio/gimbal/discovery/pkg/translator"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// returns a kubernetes service for each load balancer in the slice
func kubeServices(clusterName, tenantName string, lbs []loadbalancers.LoadBalancer) []v1.Service {
	var svcs []v1.Service
	for _, lb := range lbs {
		svc := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: tenantName,
				Name:      translator.BuildKubernetesDNSLabel(serviceName(lb), clusterName),
				Labels:    translator.AddGimbalLabels(clusterName, tenantName, serviceName(lb), loadbalancerLabels(lb)),
			},
			Spec: v1.ServiceSpec{
				Type:      v1.ServiceTypeClusterIP,
				ClusterIP: "None",
			},
		}
		for _, l := range lb.Listeners {
			svc.Spec.Ports = append(svc.Spec.Ports, servicePort(&l))
		}
		svcs = append(svcs, svc)
	}
	return svcs
}

// returns a kubernetes endpoints resource for each load balancer in the slice
func kubeEndpoints(clusterName, tenantName string, lbs []loadbalancers.LoadBalancer, ps []pools.Pool) []v1.Endpoints {
	var endpoints []v1.Endpoints
	for _, lb := range lbs {
		ep := v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: tenantName,
				Name:      translator.BuildKubernetesDNSLabel(serviceName(lb), clusterName),
				Labels:    translator.AddGimbalLabels(clusterName, tenantName, serviceName(lb), loadbalancerLabels(lb)),
			},
		}
		for _, l := range lb.Listeners {
			// compute endpoint susbsets for each listener
			subsets := map[int]v1.EndpointSubset{}

			// get the listeners pool
			var pool pools.Pool
			for _, p := range ps {
				if l.DefaultPoolID == p.ID {
					pool = p
					break
				}
			}

			// We want to group all members that are listening on the same port
			// into a single EndpointSubset. We achieve this by using a map of
			// subsets, keyed by the listening port.
			for _, member := range pool.Members {
				s := subsets[member.ProtocolPort]
				// Add the port if we haven't added it yet to the EndpointSubset
				if len(s.Ports) == 0 {
					s.Ports = append(s.Ports, v1.EndpointPort{Name: portName(&l), Port: int32(member.ProtocolPort), Protocol: v1.ProtocolTCP})
				}
				s.Addresses = append(s.Addresses, v1.EndpointAddress{IP: member.Address}) // TODO: can address be something other than an IP address?
				subsets[member.ProtocolPort] = s
			}

			// Add the subsets to the Endpoint
			for _, s := range subsets {
				ep.Subsets = append(ep.Subsets, s)
			}
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints
}

func loadbalancerLabels(lb loadbalancers.LoadBalancer) map[string]string {
	return map[string]string{
		"gimbal.heptio.com/load-balancer-id":   lb.ID,
		"gimbal.heptio.com/load-balancer-name": lb.Name,
	}
}

// By default, the load balancer's ID is used as the service name, given that
// names in OpenStack are optional and are not guaranteed to be unique. In the
// case that the load balancer has a non-empty name, the ID is appended to
// produce a unique service name.
func serviceName(lb loadbalancers.LoadBalancer) string {
	if lb.Name == "" {
		return lb.ID
	}
	return fmt.Sprintf("%s-%s", lb.Name, lb.ID)
}

func servicePort(listener *listeners.Listener) v1.ServicePort {
	pn := portName(listener)
	return v1.ServicePort{
		Name:     pn,
		Port:     int32(listener.ProtocolPort),
		Protocol: v1.ProtocolTCP, // only support TCP
	}
}

func portName(listener *listeners.Listener) string {
	p := strconv.Itoa(listener.ProtocolPort)
	if listener.Name == "" {
		return "unnamed-" + p // TODO: port names must have at least 1 char. Is there something better we can do here?
	}
	// port names must be kubernetes DNS_LABELs
	return translator.BuildKubernetesDNSLabel(listener.Name, p)
}
