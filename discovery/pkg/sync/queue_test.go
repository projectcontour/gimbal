package sync

import (
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestQueueSuccessfulSync(t *testing.T) {
	client := fake.NewSimpleClientset()
	var createAttempts int
	client.PrependReactor("create", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		createAttempts++
		return true, nil, nil
	})
	q := NewQueue(logrus.New(), client, 4, metrics.NewMetrics("test", "backend"))
	stop := make(chan struct{})
	go q.Run(stop)

	q.Enqueue(AddServiceAction(&v1.Service{}))
	// TODO(abrand): replace sleeps with some other signal
	time.Sleep(1 * time.Second)
	close(stop)

	assert.Equal(t, 1, createAttempts)
	assert.Equal(t, 0, q.Workqueue.Len())
}

func TestQueueStopsRetryingAfterSuccess(t *testing.T) {
	client := fake.NewSimpleClientset()
	var createAttempts int
	client.PrependReactor("create", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		err := errors.New("fake error")
		if createAttempts == 1 {
			err = nil
		}
		createAttempts++
		return true, nil, err
	})

	// Start the queue
	q := NewQueue(logrus.New(), client, 4, metrics.NewMetrics("test", "backend"))
	stop := make(chan struct{})
	go q.Run(stop)

	// Enqueue an add service that will always fail
	q.Enqueue(AddServiceAction(&v1.Service{}))

	// Wait until we process it
	// TODO(abrand): replace sleeps with some other signal
	time.Sleep(1 * time.Second)
	close(stop)

	// Assert that we tried twice
	assert.Equal(t, 2, createAttempts)
	assert.Equal(t, 0, q.Workqueue.Len())
}

func TestQueueMaxRetries(t *testing.T) {
	client := fake.NewSimpleClientset()
	var createAttempts int
	client.PrependReactor("create", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		createAttempts++
		return true, nil, errors.New("fake error")
	})

	// Start the queue
	q := NewQueue(logrus.New(), client, 4, metrics.NewMetrics("test", "backend"))
	stop := make(chan struct{})
	go q.Run(stop)

	// Enqueue an add service that will always fail
	q.Enqueue(AddServiceAction(&v1.Service{}))

	// Wait until we process it
	// TODO(abrand): replace sleeps with some other signal
	time.Sleep(1 * time.Second)
	close(stop)

	// Assert that we tried five times and that we finally dropped it
	assert.Equal(t, queueMaxRetries, createAttempts)
	assert.Equal(t, 0, q.Workqueue.Len())
}

func TestQueueServicesMetrics(t *testing.T) {
	now := time.Date(2000, 1, 1, 10, 0, 00, 0, time.UTC)
	tests := []struct {
		name                   string
		apiServerError         error
		expectedTimestampGauge float64
		expectedErrorCounter   float64
	}{
		{
			name: "successfull replication",
			expectedTimestampGauge: float64(now.Unix()),
			expectedErrorCounter:   float64(-1), // no error, so error counter is not initialized
		},
		{
			name: "failed to replicate service",
			expectedTimestampGauge: float64(-1), // failed to sync resource, so timestamp is not initialized
			expectedErrorCounter:   float64(queueMaxRetries),
			apiServerError:         errors.New("api server error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nowFunc = func() time.Time {
				return now
			}
			client := fake.NewSimpleClientset()
			client.PrependReactor("create", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.apiServerError
			})
			m := metrics.NewMetrics("test", "backend")
			m.RegisterPrometheus(false)
			q := NewQueue(logrus.New(), client, 1, m)

			stop := make(chan struct{})
			go q.Run(stop)

			s := v1.Service{}
			s.Name = "foo"
			s.Namespace = "default"
			q.Enqueue(AddServiceAction(&s))

			time.Sleep(500 * time.Millisecond)
			close(stop)

			assertGaugeEqual(t, test.expectedTimestampGauge, metrics.ServiceEventTimestampGauge, m.Registry)
			assertCounterEqual(t, test.expectedErrorCounter, metrics.ServiceErrorTotalCounter, m.Registry)
		})
	}
}

func TestQueueEndpointsMetrics(t *testing.T) {
	now := time.Date(2000, 1, 1, 10, 0, 00, 0, time.UTC)

	tests := []struct {
		name                             string
		apiServerError                   error
		expectedTimestampGauge           float64
		expectedErrorCounter             float64
		expectedReplicatedEndpointsGauge float64
	}{
		{
			name: "successfull replication",
			expectedTimestampGauge:           float64(now.Unix()),
			expectedErrorCounter:             float64(-1), // no error, so error counter is not initialized
			expectedReplicatedEndpointsGauge: float64(2),
		},
		{
			name: "failed to replicate endpoints resource",
			expectedTimestampGauge:           float64(-1), // failed to sync resource, so timestamp is not initialized
			expectedErrorCounter:             float64(queueMaxRetries),
			expectedReplicatedEndpointsGauge: float64(-1), // failed to replicate, so gauge is not initialized
			apiServerError:                   errors.New("api server error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nowFunc = func() time.Time {
				return now
			}
			client := fake.NewSimpleClientset()
			client.PrependReactor("create", "endpoints", func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.apiServerError
			})
			m := metrics.NewMetrics("test", "backend")
			m.RegisterPrometheus(false)
			q := NewQueue(logrus.New(), client, 1, m)

			stop := make(chan struct{})
			go q.Run(stop)

			ep := v1.Endpoints{}
			ep.Namespace = "default"
			ep.Name = "foo"
			ep.Subsets = []v1.EndpointSubset{{
				Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}, {IP: "10.0.0.2"}},
			}}
			q.Enqueue(AddEndpointsAction(&ep, "upstream"))

			time.Sleep(500 * time.Millisecond)
			close(stop)

			assertGaugeEqual(t, test.expectedTimestampGauge, metrics.EndpointsEventTimestampGauge, m.Registry)
			assertCounterEqual(t, test.expectedErrorCounter, metrics.EndpointsErrorTotalCounter, m.Registry)
			assertGaugeEqual(t, test.expectedReplicatedEndpointsGauge, metrics.DiscovererReplicatedEndpointsGauge, m.Registry)
		})
	}
}

func assertGaugeEqual(t *testing.T, expected float64, metricName string, reg *prometheus.Registry) {
	mf, err := reg.Gather()
	require.NoError(t, err, "gathering metrics")
	v := float64(-1)
	for _, m := range mf {
		if m.GetName() == metricName {
			v = m.GetMetric()[0].GetGauge().GetValue()
		}
	}
	assert.Equal(t, expected, v)
}

func assertCounterEqual(t *testing.T, expected float64, metricName string, reg *prometheus.Registry) {
	mf, err := reg.Gather()
	require.NoError(t, err, "gathering metrics")
	v := float64(-1)
	for _, m := range mf {
		if m.GetName() == metricName {
			v = m.GetMetric()[0].GetCounter().GetValue()
		}
	}
	assert.Equal(t, expected, v)
}
