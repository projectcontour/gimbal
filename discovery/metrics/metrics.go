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
	Metrics map[string]prometheus.Collector
}

const (
	ServiceEventTimestampGauge     = "service_event_timestamp"
	EndpointsEventTimestampGauge   = "endpoints_event_timestamp"
	ServiceErrorTotalCounter       = "service_error_total"
	EndpointsErrorTotalCounter     = "endpoints_error_total"
	QuesizeGauge                   = "queuesize"
	DiscovererAPILatencyMSGauge    = "discoverer_api_latency_ms"
	DiscovererCycleDurationMSGauge = "discoverer_cycle_duration_ms"
)

// NewMetrics returns a map of Prometheus metrics
func NewMetrics() DiscovererMetrics {

	metrics := make(map[string]prometheus.Collector)
	metrics[ServiceEventTimestampGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gimbal_service_event_timestamp",
			Help: "Timestamp last service event was processed",
		},
		[]string{"namespace", "clustername", "name"},
	)

	metrics[EndpointsEventTimestampGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gimbal_endpoints_event_timestamp",
			Help: "Timestamp last endpoints event was processed",
		},
		[]string{"namespace", "clustername", "name"},
	)

	metrics[ServiceErrorTotalCounter] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gimbal_service_error_total",
			Help: "Number of service errors encountered",
		},
		[]string{"namespace", "clustername", "name", "errortype"},
	)

	metrics[EndpointsErrorTotalCounter] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gimbal_endpoints_error_total",
			Help: "Number of endpoints errors encountered",
		},
		[]string{"namespace", "clustername", "name", "errortype"},
	)

	metrics[QuesizeGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gimbal_queuesize",
			Help: "Number of items in process queue",
		},
		[]string{"namespace", "clustername", "clustertype"},
	)

	metrics[DiscovererAPILatencyMSGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gimbal_discoverer_api_latency_ms",
			Help: "The milliseconds it takes for requests to return from a remote discoverer api",
		},
		[]string{"clustername", "clustertype"},
	)

	metrics[DiscovererCycleDurationMSGauge] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gimbal_discoverer_cycle_duration_ms",
			Help: "The milliseconds it takes for all objects to be synced from a remote discoverer api",
		},
		[]string{"clustername", "clustertype"},
	)

	return DiscovererMetrics{
		Metrics: metrics,
	}
}

// ServiceMetricError formats a prometheus metric and increments
func (d *DiscovererMetrics) ServiceMetricError(namespace, clusterName, serviceName, errtype string) {
	m, ok := d.Metrics[ServiceErrorTotalCounter].(*prometheus.CounterVec)
	if ok {
		m.WithLabelValues(namespace, clusterName, serviceName, errtype).Inc()
	}
}

// EndpointsMetricError formats a prometheus metric and increments
func (d *DiscovererMetrics) EndpointsMetricError(namespace, clusterName, endpointsName, errtype string) {
	m, ok := d.Metrics[EndpointsErrorTotalCounter].(*prometheus.CounterVec)
	if ok {
		m.WithLabelValues(namespace, clusterName, endpointsName, errtype).Inc()
	}
}
