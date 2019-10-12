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

package openstack

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/pools"
	"github.com/vmware-tanzu/gimbal/discovery/pkg/translator"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// returns a kubernetes service for each load balancer in the slice
func kubeServices(backendName, tenantName string, lbs []loadbalancers.LoadBalancer) []v1.Service {
	var svcs []v1.Service
	for _, lb := range lbs {
		svc := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: tenantName,
				Name:      translator.BuildDiscoveredName(backendName, serviceName(lb)),
				Labels:    translator.AddGimbalLabels(backendName, serviceName(lb), loadbalancerLabels(lb)),
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
func kubeEndpoints(backendName, tenantName string, lbs []loadbalancers.LoadBalancer, ps []pools.Pool) []Endpoints {
	endpoints := []Endpoints{}
	for _, lb := range lbs {
		ep := v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: tenantName,
				Name:      translator.BuildDiscoveredName(backendName, serviceName(lb)),
				Labels:    translator.AddGimbalLabels(backendName, serviceName(lb), loadbalancerLabels(lb)),
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
		endpoints = append(endpoints, Endpoints{endpoints: ep, upstreamName: serviceNameOriginal(lb)})
	}
	return endpoints
}

func loadbalancerLabels(lb loadbalancers.LoadBalancer) map[string]string {
	// Sanitize the load balancer name according to the kubernetes label value
	// requirements: "Valid label values must be 63 characters or less and must
	// be empty or begin and end with an alphanumeric character ([a-z0-9A-Z])
	// with dashes (-), underscores (_), dots (.), and alphanumerics between."
	name := lb.Name
	if name != "" {
		// 1. replace unallowed chars with a dash
		reg := regexp.MustCompile("[^a-zA-Z0-9\\-._]")
		name = reg.ReplaceAllString(lb.Name, "-")

		// 2. prepend/append a special marker if first/last char is not an alphanum
		if !isalphanum(name[0]) {
			name = "lb" + name
		}
		if !isalphanum(name[len(name)-1]) {
			name = name + "lb"
		}
		// 3. shorten if necessary
		name = translator.ShortenKubernetesLabelValue(name)
	}
	return map[string]string{
		"gimbal.heptio.com/load-balancer-id":   lb.ID,
		"gimbal.heptio.com/load-balancer-name": name,
	}
}

func isalphanum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// use the load balancer ID as the service name
// context: heptio/gimbal #216
func serviceName(lb loadbalancers.LoadBalancer) string {
	return strings.ToLower(lb.ID)
}

// get the lb Name or ID if name is empty
func serviceNameOriginal(lb loadbalancers.LoadBalancer) string {
	lbName := lb.Name
	if lb.Name == "" {
		lbName = lb.ID
	}
	return strings.ToLower(lbName)
}

func servicePort(listener *listeners.Listener) v1.ServicePort {
	pn := portName(listener)
	return v1.ServicePort{
		Name: pn,
		Port: int32(listener.ProtocolPort),
		// The K8s API server sets this field on service creation. By setting
		// this ourselves, we prevent the discoverer from thinking it needs to
		// perform an update every time it compares the translated object with
		// the one that exists in gimbal.
		TargetPort: intstr.FromInt(listener.ProtocolPort),
		Protocol:   v1.ProtocolTCP, // only support TCP
	}
}

func portName(listener *listeners.Listener) string {
	p := strconv.Itoa(listener.ProtocolPort)
	return "port-" + p
}
