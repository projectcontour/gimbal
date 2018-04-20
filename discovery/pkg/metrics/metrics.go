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
	"github.com/prometheus/client_golang/prometheus"
)

// DiscovererMetrics provide Prometheus metrics for the app
type DiscovererMetrics struct {
	metrics map[string]prometheus.Collector
}

const (
	ServiceEventTimestampGauge     = "gimbal_service_event_timestamp"
	EndpointsEventTimestampGauge   = "gimbal_endpoints_event_timestamp"
	ServiceErrorTotalCounter       = "gimbal_service_error_total"
	EndpointsErrorTotalCounter     = "gimbal_endpoints_error_total"
	QueueSizeGauge                 = "gimbal_queuesize"
	DiscovererAPILatencyMSGauge    = "gimbal_discoverer_api_latency_ms"
	DiscovererCycleDurationMSGauge = "gimbal_discoverer_cycle_duration_ms"
	DiscovererErrorTotal           = "gimbal_discoverer_error_total"
)

// NewMetrics returns a map of Prometheus metrics
func NewMetrics() DiscovererMetrics {

	metrics := make(map[string]prometheus.Collector)
	metrics[ServiceEventTimestampGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: ServiceEventTimestampGauge,
			Help: "Timestamp last service event was processed",
		},
		[]string{"namespace", "clustername", "name"},
	)

	metrics[EndpointsEventTimestampGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: EndpointsEventTimestampGauge,
			Help: "Timestamp last endpoints event was processed",
		},
		[]string{"namespace", "clustername", "name"},
	)

	metrics[ServiceErrorTotalCounter] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: ServiceErrorTotalCounter,
			Help: "Number of service errors encountered",
		},
		[]string{"namespace", "clustername", "name", "errortype"},
	)

	metrics[EndpointsErrorTotalCounter] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: EndpointsErrorTotalCounter,
			Help: "Number of endpoints errors encountered",
		},
		[]string{"namespace", "clustername", "name", "errortype"},
	)

	metrics[QueueSizeGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: QueueSizeGauge,
			Help: "Number of items in process queue",
		},
		[]string{"namespace", "clustername", "clustertype"},
	)

	metrics[DiscovererAPILatencyMSGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: DiscovererAPILatencyMSGauge,
			Help: "The milliseconds it takes for requests to return from a remote discoverer api",
		},
		[]string{"clustername", "clustertype"},
	)

	metrics[DiscovererCycleDurationMSGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: DiscovererCycleDurationMSGauge,
			Help: "The milliseconds it takes for all objects to be synced from a remote discoverer api",
		},
		[]string{"clustername", "clustertype"},
	)

	metrics[DiscovererErrorTotal] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: DiscovererErrorTotal,
			Help: "Number of errors that have occurred in the Discoverer",
		},
		[]string{"clustername", "errortype"},
	)

	return DiscovererMetrics{
		metrics: metrics,
	}
}

// RegisterPrometheus registers the metrics
func (d *DiscovererMetrics) RegisterPrometheus() {
	// Register with Prometheus's default registry
	for _, v := range d.metrics {
		prometheus.MustRegister(v)
	}
}

// ServiceMetricError formats a service prometheus metric and increments
func (d *DiscovererMetrics) ServiceMetricError(namespace, clusterName, serviceName, errtype string) {
	m, ok := d.metrics[ServiceErrorTotalCounter].(*prometheus.CounterVec)
	if ok {
		m.WithLabelValues(namespace, clusterName, serviceName, errtype).Inc()
	}
}

// EndpointsMetricError formats an endpoint prometheus metric and increments
func (d *DiscovererMetrics) EndpointsMetricError(namespace, clusterName, endpointsName, errtype string) {
	m, ok := d.metrics[EndpointsErrorTotalCounter].(*prometheus.CounterVec)
	if ok {
		m.WithLabelValues(namespace, clusterName, endpointsName, errtype).Inc()
	}
}

// GenericMetricError formats a generic prometheus metric and increments
func (d *DiscovererMetrics) GenericMetricError(clusterName, errtype string) {
	m, ok := d.metrics[DiscovererErrorTotal].(*prometheus.CounterVec)
	if ok {
		m.WithLabelValues(clusterName, errtype).Inc()
	}
}

// ServiceEventTimestampMetric formats a Service event timestamp prometheus metric
func (d *DiscovererMetrics) ServiceEventTimestampMetric(namespace, clusterName, name string, timestamp int64) {
	m, ok := d.metrics[ServiceEventTimestampGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(namespace, clusterName, name).Set(float64(timestamp))
	}
}

// EndpointsEventTimestampMetric formats a Endpoint event timestamp prometheus metric
func (d *DiscovererMetrics) EndpointsEventTimestampMetric(namespace, clusterName, name string, timestamp int64) {
	m, ok := d.metrics[EndpointsEventTimestampGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(namespace, clusterName, name).Set(float64(timestamp))
	}
}

// QueueSizeGaugeMetric records the queue size prometheus metric
func (d *DiscovererMetrics) QueueSizeGaugeMetric(namespace, clusterName, clusterType string, size int) {
	m, ok := d.metrics[QueueSizeGauge].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(namespace, clusterName, clusterType).Set(float64(size))
	}
}
