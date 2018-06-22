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

package metrics

import (
	"math"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// DiscovererMetrics provide Prometheus metrics for the app
type DiscovererMetrics struct {
	Registry    *prometheus.Registry
	Metrics     map[string]prometheus.Collector
	BackendType string
	BackendName string
}

const (
	ServiceEventTimestampGauge              = "gimbal_service_event_timestamp"
	EndpointsEventTimestampGauge            = "gimbal_endpoints_event_timestamp"
	ServiceErrorTotalCounter                = "gimbal_service_error_total"
	EndpointsErrorTotalCounter              = "gimbal_endpoints_error_total"
	QueueSizeGauge                          = "gimbal_queuesize"
	DiscovererAPILatencyMsHistogram         = "gimbal_discoverer_api_latency_milliseconds"
	DiscovererCycleDurationSecondsHistogram = "gimbal_discoverer_cycle_duration_seconds"
	DiscovererErrorTotal                    = "gimbal_discoverer_error_total"
	DiscovererUpstreamServicesGauge         = "gimbal_discoverer_upstream_services_total"
	DiscovererReplicatedServicesGauge       = "gimbal_discoverer_replicated_services_total"
	DiscovererInvalidServicesGauge          = "gimbal_discoverer_invalid_services_total"
	DiscovererUpstreamEndpointsGauge        = "gimbal_discoverer_upstream_endpoints_total"
	DiscovererReplicatedEndpointsGauge      = "gimbal_discoverer_replicated_endpoints_total"
	DiscovererInvalidEndpointsGauge         = "gimbal_discoverer_invalid_endpoints_total"
	DiscovererInfoGauge                     = "gimbal_discoverer_info"
)

// NewMetrics returns a map of Prometheus metrics
func NewMetrics(BackendType, BackendName string) DiscovererMetrics {
	return DiscovererMetrics{
		Registry:    prometheus.NewRegistry(),
		BackendType: BackendType,
		BackendName: BackendName,
		Metrics: map[string]prometheus.Collector{
			ServiceEventTimestampGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: ServiceEventTimestampGauge,
					Help: "Timestamp last service event was processed",
				},
				[]string{"namespace", "backendname", "name", "backendtype"},
			),
			EndpointsEventTimestampGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: EndpointsEventTimestampGauge,
					Help: "Timestamp last endpoints event was processed",
				},
				[]string{"namespace", "backendname", "name", "backendtype"},
			),
			ServiceErrorTotalCounter: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: ServiceErrorTotalCounter,
					Help: "Number of service errors encountered",
				},
				[]string{"namespace", "backendname", "name", "errortype", "backendtype"},
			),
			EndpointsErrorTotalCounter: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: EndpointsErrorTotalCounter,
					Help: "Number of endpoints errors encountered",
				},
				[]string{"namespace", "backendname", "name", "errortype", "backendtype"},
			),
			QueueSizeGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: QueueSizeGauge,
					Help: "Number of items in process queue",
				},
				[]string{"backendname", "backendtype"},
			),
			DiscovererAPILatencyMsHistogram: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    DiscovererAPILatencyMsHistogram,
					Help:    "The milliseconds it takes for requests to return from a remote discoverer api",
					Buckets: []float64{20, 50, 100, 250, 500, 1000, 2000, 5000, 10000, 20000, 50000, 120000}, // milliseconds. largest bucket is 2 minutes.
				},
				[]string{"backendname", "backendtype", "path"},
			),
			DiscovererCycleDurationSecondsHistogram: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    DiscovererCycleDurationSecondsHistogram,
					Help:    "The seconds it takes for all objects to be synced from a remote backend",
					Buckets: prometheus.LinearBuckets(60, 60, 10), // 10 buckets, each 30 wide
				},
				[]string{"backendname", "backendtype"},
			),
			DiscovererErrorTotal: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: DiscovererErrorTotal,
					Help: "Number of errors that have occurred in the Discoverer",
				},
				[]string{"backendname", "errortype", "backendtype"},
			),
			DiscovererUpstreamServicesGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: DiscovererUpstreamServicesGauge,
					Help: "Total number of services in the backend",
				},
				[]string{"backendname", "namespace", "backendtype"},
			),
			DiscovererReplicatedServicesGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: DiscovererReplicatedServicesGauge,
					Help: "Total number of services replicated from the backend",
				},
				[]string{"backendname", "namespace", "backendtype"},
			),
			DiscovererInvalidServicesGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: DiscovererInvalidServicesGauge,
					Help: "Total number of invalid services that could not be replicated from the backend",
				},
				[]string{"backendname", "namespace", "backendtype"},
			),
			DiscovererUpstreamEndpointsGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: DiscovererUpstreamEndpointsGauge,
					Help: "Total number of endpoints in the backend",
				},
				[]string{"backendname", "namespace", "servicename", "backendtype"},
			),
			DiscovererReplicatedEndpointsGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: DiscovererReplicatedEndpointsGauge,
					Help: "Total number of endpoints replicated the backend",
				},
				[]string{"backendname", "namespace", "servicename", "backendtype"},
			),
			DiscovererInvalidEndpointsGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: DiscovererInvalidEndpointsGauge,
					Help: "Total number of invalid endpoints that could not be replicated from the backend",
				},
				[]string{"backendname", "namespace", "servicename", "backendtype"},
			),
			DiscovererInfoGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: DiscovererInfoGauge,
					Help: "Information about discoverer",
				},
				[]string{"backendname", "version", "backendtype"},
			),
		},
	}
}

// RegisterPrometheus registers the Metrics
func (d *DiscovererMetrics) RegisterPrometheus(registerDefault bool) {

	if registerDefault {
		// Register detault process / go collectors
		d.Registry.MustRegister(prometheus.NewProcessCollector(os.Getpid(), ""))
		d.Registry.MustRegister(prometheus.NewGoCollector())
	}

	// Register with Prometheus's default registry
	for _, v := range d.Metrics {
		d.Registry.MustRegister(v)
	}
}

// ServiceMetricError formats a service prometheus metric and increments
func (d *DiscovererMetrics) ServiceMetricError(namespace, serviceName, errtype string) {
	m, ok := d.Metrics[ServiceErrorTotalCounter].(*prometheus.CounterVec)
	if ok {
		m.WithLabelValues(namespace, d.BackendName, serviceName, errtype, d.BackendType).Inc()
	}
}

// EndpointsMetricError formats an endpoint prometheus metric and increments
func (d *DiscovererMetrics) EndpointsMetricError(namespace, endpointsName, errtype string) {
	m, ok := d.Metrics[EndpointsErrorTotalCounter].(*prometheus.CounterVec)
	if ok {
		m.WithLabelValues(namespace, d.BackendName, endpointsName, errtype, d.BackendType).Inc()
	}
}

// GenericMetricError formats a generic prometheus metric and increments
func (d *DiscovererMetrics) GenericMetricError(errtype string) {
	m, ok := d.Metrics[DiscovererErrorTotal].(*prometheus.CounterVec)
	if ok {
		m.WithLabelValues(d.BackendName, errtype, d.BackendType).Inc()
	}
}

// ServiceEventTimestampMetric formats a Service event timestamp prometheus metric
func (d *DiscovererMetrics) ServiceEventTimestampMetric(namespace, name string, timestamp int64) {
	m, ok := d.Metrics[ServiceEventTimestampGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(namespace, d.BackendName, name, d.BackendType).Set(float64(timestamp))
	}
}

// EndpointsEventTimestampMetric formats a Endpoint event timestamp prometheus metric
func (d *DiscovererMetrics) EndpointsEventTimestampMetric(namespace, name string, timestamp int64) {
	m, ok := d.Metrics[EndpointsEventTimestampGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(namespace, d.BackendName, name, d.BackendType).Set(float64(timestamp))
	}
}

// QueueSizeGaugeMetric records the queue size prometheus metric
func (d *DiscovererMetrics) QueueSizeGaugeMetric(size int) {
	m, ok := d.Metrics[QueueSizeGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(d.BackendName, d.BackendType).Set(float64(size))
	}
}

// CycleDurationMetric formats a cycle duration gauge prometheus metric
func (d *DiscovererMetrics) CycleDurationMetric(duration time.Duration) {
	m, ok := d.Metrics[DiscovererCycleDurationSecondsHistogram].(*prometheus.HistogramVec)
	if ok {
		m.WithLabelValues(d.BackendName, d.BackendType).Observe(math.Floor(duration.Seconds()))
	}
}

// APILatencyMetric formats a cycle duration gauge prometheus metric
func (d *DiscovererMetrics) APILatencyMetric(path string, duration time.Duration) {
	m, ok := d.Metrics[DiscovererAPILatencyMsHistogram].(*prometheus.HistogramVec)
	if ok {
		m.WithLabelValues(d.BackendName, d.BackendType, path).Observe(math.Floor(duration.Seconds() * 1e3))
	}
}

// DiscovererUpstreamServicesMetric records the total number of upstream services
func (d *DiscovererMetrics) DiscovererUpstreamServicesMetric(namespace string, totalServices int) {
	m, ok := d.Metrics[DiscovererUpstreamServicesGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(d.BackendName, namespace, d.BackendType).Set(float64(totalServices))
	}
}

// DiscovererReplicatedServicesMetric records the total number of replicated services
func (d *DiscovererMetrics) DiscovererReplicatedServicesMetric(namespace string, totalServices int) {
	m, ok := d.Metrics[DiscovererReplicatedServicesGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(d.BackendName, namespace, d.BackendType).Set(float64(totalServices))
	}
}

// DiscovererInvalidServicesMetric records the total number of invalid services
func (d *DiscovererMetrics) DiscovererInvalidServicesMetric(namespace string, totalServices int) {
	m, ok := d.Metrics[DiscovererInvalidServicesGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(d.BackendName, namespace, d.BackendType).Set(float64(totalServices))
	}
}

// DiscovererUpstreamEndpointsMetric records the total upstream endpoints in the backend
func (d *DiscovererMetrics) DiscovererUpstreamEndpointsMetric(namespace, serviceName string, totalEp int) {
	m, ok := d.Metrics[DiscovererUpstreamEndpointsGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(d.BackendName, namespace, serviceName, d.BackendType).Set(float64(totalEp))
	}
}

// DiscovererReplicatedEndpointsMetric records the total replicated endpoints
func (d *DiscovererMetrics) DiscovererReplicatedEndpointsMetric(namespace, serviceName string, totalEp int) {
	m, ok := d.Metrics[DiscovererReplicatedEndpointsGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(d.BackendName, namespace, serviceName, d.BackendType).Set(float64(totalEp))
	}
}

// DiscovererInfoMetric records version information
func (d *DiscovererMetrics) DiscovererInfoMetric(version string) {
	m, ok := d.Metrics[DiscovererInfoGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(d.BackendName, version, d.BackendType).Set(1)
	}
}
