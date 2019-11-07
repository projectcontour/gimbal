package vmware

import (
	"strconv"

	"github.com/projectcontour/gimbal/pkg/translator"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// returns a kubernetes service
func kubeServices(backendName string, svcs []VirtualMachine) []v1.Service {

	var services []v1.Service
	for _, vm := range svcs {

		intPort, _ := strconv.Atoi(vm.Port) // TODO(sas): Handle the error for conversion!

		svc := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: vm.ProjectName,
				Name:      translator.BuildDiscoveredName(backendName, vm.Application),
				Labels:    translator.AddGimbalLabels(backendName, vm.Application, nil),
			},
			Spec: v1.ServiceSpec{
				Type:      v1.ServiceTypeClusterIP,
				ClusterIP: "None",
				Ports: []v1.ServicePort{{
					Name:     vm.Application,
					Protocol: "TCP",
					Port:     int32(intPort),
				}},
			},
		}
		services = append(services, svc)
	}
	return services
}

// returns a kubernetes endpoints resource for each load balancer in the slice
func kubeEndpoints(backendName string, namespace, application string, eps []VirtualMachine) []translator.Endpoint {

	// compute endpoint susbsets
	subsets := map[int]v1.EndpointSubset{}

	ep := v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      translator.BuildDiscoveredName(backendName, application),
			Labels:    translator.AddGimbalLabels(backendName, application, nil),
		},
	}
	for _, vm := range eps {

		intPort, _ := strconv.Atoi(vm.Port)

		// We want to group all members that are listening on the same port
		// into a single EndpointSubset. We achieve this by using a map of
		// subsets, keyed by the listening port.
		s := subsets[intPort]
		// Add the port if we haven't added it yet to the EndpointSubset
		if len(s.Ports) == 0 {
			s.Ports = append(s.Ports, v1.EndpointPort{Name: vm.Application, Port: int32(intPort), Protocol: v1.ProtocolTCP})
		}
		s.Addresses = append(s.Addresses, v1.EndpointAddress{IP: vm.IPAddress}) // TODO: can address be something other than an IP address?
		subsets[intPort] = s

		// Add the subsets to the Endpoint
		for _, s := range subsets {
			ep.Subsets = append(ep.Subsets, s)
		}
	}
	return []translator.Endpoint{{
		Endpoints:    ep,
		UpstreamName: application,
	}}
}
