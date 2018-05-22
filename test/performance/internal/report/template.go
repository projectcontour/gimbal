package report

const prettyReportTmpl = `
Gimbalbench Report
Report generated: {{.Generated.Format "Mon 02 Jan 2006 15:04:05 MST" }}
Test start time: {{.TestConfiguration.StartTime.Format "Mon 02 Jan 2006 15:04:05 MST" }}

Test configuration:
Wrk2 Nodes: {{.TestConfiguration.WrkNodeCount}}
Nginx Nodes: {{.TestConfiguration.NginxNodeCount}}
Wrk2 Request Rate: {{.TestConfiguration.WrkRequestRate}}
Wrk2 Host Network: {{.TestConfiguration.WrkHostNetwork}}

{{if .ConcurrentConnections -}}
Concurrent Connections Test Results:

Connections,P99 Latency, P999 Latency
{{- range .ConcurrentConnections}}
{{.Connections}},{{.Latency.P99}},{{.Latency.P999}}
{{- end}}
{{- end}}

{{if .BackendServices}}
Backend Services Count Test Results:

# of Services,P99 Latency, P999 Latency
{{- range .BackendServices}}
{{.BackendServices}},{{.Latency.P99}},{{.Latency.P999}}
{{- end}}
{{- end}}

{{if .BackendEndpointsOnSingleService}}
Backend Endpoints on Single Service Test Results:

# of Endpoints, P99 Latency, P999 Latency
{{- range .BackendEndpointsOnSingleService}}
{{.Endpoints}},{{.Latency.P99}},{{.Latency.P999}}
{{- end}}
{{- end}}

{{if .IngressCount}}
Ingress Count Test Results:

# of Ingresses, P99 Latency, P999 Latency
{{- range .IngressCount}}
{{.IngressResources}},{{.Latency.P99}},{{.Latency.P999}}
{{- end}}
{{- end}}

{{if .IngressRoutes}}
IngressRoutes Test Results:

# of IngressRoutes, P99 Latency, P999 Latency
{{- range .IngressRoutes}}
{{.IngressRoutes}},{{.Latency.P99}},{{.Latency.P999}}
{{- end}}
{{- end}}

{{if .KubernetesDiscoveryTime }}
Kubernetes Discovery Time Results:
{{- range .KubernetesDiscoveryTime }}
# of services, Time to 1st discovery, Time to full discovery, Time to discover new, Time to discover update
{{.ServiceCount}},{{.TimeToFirstDiscovery}},{{.TimeToFullDiscovery}},{{.TimeToDiscoverNew}},{{.TimeToDiscoverUpdate}}
{{- end}}
{{- end}}
`
