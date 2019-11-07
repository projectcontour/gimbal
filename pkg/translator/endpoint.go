package translator

import v1 "k8s.io/api/core/v1"

// Endpoint represents a v1.Endpoint + upstream name to facilitate metrics
type Endpoint struct {
	Endpoints    v1.Endpoints
	UpstreamName string
}
