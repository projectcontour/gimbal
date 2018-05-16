package openstack

import (
	"testing"

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
