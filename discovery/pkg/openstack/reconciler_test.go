package openstack

import (
	"fmt"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSkipInvalidLoadBalancers(t *testing.T) {
	tests := []struct {
		name     string
		lbs      []loadbalancers.LoadBalancer
		expected []loadbalancers.LoadBalancer
	}{
		{
			name:     "empty slice",
			lbs:      []loadbalancers.LoadBalancer{},
			expected: []loadbalancers.LoadBalancer{},
		},
		{
			name: "unnamed lb",
			lbs: []loadbalancers.LoadBalancer{
				{ID: "foobarbaz"},
			},
			expected: []loadbalancers.LoadBalancer{
				{ID: "foobarbaz"},
			},
		},
		{
			name: "valid lbs",
			lbs: []loadbalancers.LoadBalancer{
				{Name: "foobarbaz"},
				{Name: "foo123"},
				{Name: "123foo"},
				{Name: "foo-bar"},
				{Name: "foo-bar-123"},
				{Name: "1-2-3-4-5"},
				{Name: "FOObar"},
				{Name: "fooBar"},
			},
			expected: []loadbalancers.LoadBalancer{
				{Name: "foobarbaz"},
				{Name: "foo123"},
				{Name: "123foo"},
				{Name: "foo-bar"},
				{Name: "foo-bar-123"},
				{Name: "1-2-3-4-5"},
				{Name: "FOObar"},
				{Name: "fooBar"},
			},
		},
		{
			name: "multiple lbs, one invalid",
			lbs: []loadbalancers.LoadBalancer{
				{Name: "foo123-Bar"},
				{Name: "@123bar"},
				{Name: "bar123-Baz"},
			},
			expected: []loadbalancers.LoadBalancer{
				{Name: "foo123-Bar"},
				{Name: "bar123-Baz"},
			},
		},
		{
			name: "multiple invalid lbs",
			lbs: []loadbalancers.LoadBalancer{
				{Name: "@bar"},
				{Name: "foo@"},
				{Name: "foo_bar"},
				{Name: "foobar!"},
				{Name: "foo-bar!"},
				{Name: "----"},
				{Name: "foo-"},
			},
			expected: []loadbalancers.LoadBalancer{},
		},
	}
	for _, tc := range tests {
		r := Reconciler{
			Logger: logrus.New(),
		}
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, r.skipInvalidLoadBalancers("projectname", tc.lbs))
		})
	}
}

func TestShouldRefreshAuthToken(t *testing.T) {
	tests := []struct {
		name             string
		refreshDuration  time.Duration
		now              string
		lastTokenRefresh string
		expected         bool
	}{
		{
			name:             "should not refresh",
			refreshDuration:  time.Second * 30,
			now:              "Mon, 02 Jan 2006 15:04:05 EST",
			lastTokenRefresh: "Mon, 02 Jan 2006 15:04:05 EST",
			expected:         false,
		},
		{
			name:             "should not refresh",
			refreshDuration:  time.Second * 30,
			now:              "Mon, 02 Jan 2006 15:04:15 EST",
			lastTokenRefresh: "Mon, 02 Jan 2006 15:04:05 EST",
			expected:         false,
		},
		{
			name:             "should refresh",
			refreshDuration:  time.Second * 30,
			now:              "Mon, 02 Jan 2006 15:04:05 EST",
			lastTokenRefresh: "Mon, 02 Jan 2006 15:03:05 EST",
			expected:         true,
		},
		{
			name:             "should not refresh",
			refreshDuration:  0,
			now:              "Mon, 02 Jan 2006 15:04:05 EST",
			lastTokenRefresh: "Mon, 01 Jan 2006 15:03:05 EST",
			expected:         false,
		},
	}

	for _, tc := range tests {

		nowParsed, err := time.Parse(time.RFC1123, tc.now)
		if err != nil {
			assert.Error(t, err)
		}

		lastTokenRefreshParsed, err := time.Parse(time.RFC1123, tc.lastTokenRefresh)
		if err != nil {
			assert.Error(t, err)
		}

		fmt.Println(lastTokenRefreshParsed.Sub(nowParsed))

		r := Reconciler{
			Logger:                 logrus.New(),
			AuthTokenRefreshPeriod: tc.refreshDuration,
			lastTokenRefresh:       lastTokenRefreshParsed,
		}

		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, r.shouldRefreshAuthToken(nowParsed))
		})
	}

}
