package sync

import (
	"errors"
	"testing"
	"time"

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
