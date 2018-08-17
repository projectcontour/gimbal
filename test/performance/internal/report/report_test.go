package report

import (
	"os"
	"testing"
	"time"
)

func TestPrettyPrintReport(t *testing.T) {
	r := Report{
		Generated: time.Now(),
		TestConfiguration: TestConfiguration{
			StartTime:      time.Now(),
			WrkRequestRate: 100,
			WrkHostNetwork: true,
			WrkNodeCount:   2,
			NginxNodeCount: 2,
		},
		ConcurrentConnections: []ConcurrentConnectionsTestResult{
			{
				Connections: 100,
				Latency: Latency{
					P99:  "1",
					P999: "1",
				},
			},
		},
		BackendServices: []BackendServicesCountTestResult{
			{
				BackendServices: 100,
				Latency: Latency{
					P99:  "1",
					P999: "1",
				},
			},
		},
		BackendEndpointsOnSingleService: []BackendEndpointsOnSingleServiceTestResult{
			{
				Endpoints: 100,
				Latency: Latency{
					P99:  "1",
					P999: "1",
				},
			},
		},
		IngressCount: []IngressCountTestResult{
			{
				IngressResources: 100,
				Latency: Latency{
					P99:  "1",
					P999: "1",
				},
			},
		},
		IngressRoutes: []IngressRouteCountTestResult{
			{
				IngressRoutes: 100,
				Latency: Latency{
					P99:  "1",
					P999: "1",
				},
			},
		},
	}
	err := PrettyPrint(os.Stdout, r)
	if err != nil {
		t.Error(err)
	}
}
