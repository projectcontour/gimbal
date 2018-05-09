package openstack

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
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
			name: "single valid lb, no listeners",
			lbs: []loadbalancers.LoadBalancer{{
				Name:      "foo123-Bar",
				Listeners: []listeners.Listener{},
			}},
			expected: []loadbalancers.LoadBalancer{{
				Name:      "foo123-Bar",
				Listeners: []listeners.Listener{},
			}},
		},
		{
			name: "single valid lb, single valid listener",
			lbs: []loadbalancers.LoadBalancer{{
				Name:      "foo123-Bar",
				Listeners: []listeners.Listener{{Name: "bar123-baz"}},
			}},
			expected: []loadbalancers.LoadBalancer{{
				Name:      "foo123-Bar",
				Listeners: []listeners.Listener{{Name: "bar123-baz"}},
			}},
		},
		{
			name: "single lb with invalid name (starts with number), no listener",
			lbs: []loadbalancers.LoadBalancer{{
				Name:      "123foo",
				Listeners: []listeners.Listener{},
			}},
			expected: []loadbalancers.LoadBalancer{},
		},
		{
			name: "single lb with invalid listener (name starts with number)",
			lbs: []loadbalancers.LoadBalancer{{
				Name:      "foo",
				Listeners: []listeners.Listener{{Name: "123foo"}},
			}},
			expected: []loadbalancers.LoadBalancer{},
		},
		{
			name: "single lb with valid and one invalid listener (name starts with number)",
			lbs: []loadbalancers.LoadBalancer{{
				Name:      "foo",
				Listeners: []listeners.Listener{{Name: "foo"}, {Name: "123foo"}, {Name: "bar"}},
			}},
			expected: []loadbalancers.LoadBalancer{},
		},
		{
			name: "multiple valid lbs, no listeners",
			lbs: []loadbalancers.LoadBalancer{
				{
					Name:      "foo123-Bar",
					Listeners: []listeners.Listener{},
				},
				{
					Name:      "bar123-Baz",
					Listeners: []listeners.Listener{},
				},
			},
			expected: []loadbalancers.LoadBalancer{
				{
					Name:      "foo123-Bar",
					Listeners: []listeners.Listener{},
				},
				{
					Name:      "bar123-Baz",
					Listeners: []listeners.Listener{},
				},
			},
		},
		{
			name: "multiple valid lbs, valid listeners",
			lbs: []loadbalancers.LoadBalancer{
				{
					Name:      "foo123-Bar",
					Listeners: []listeners.Listener{{Name: "foo123-Barlis"}},
				},
				{
					Name:      "bar123-Baz",
					Listeners: []listeners.Listener{{Name: "bar123-Bazlis"}},
				},
			},
			expected: []loadbalancers.LoadBalancer{
				{
					Name:      "foo123-Bar",
					Listeners: []listeners.Listener{{Name: "foo123-Barlis"}},
				},
				{
					Name:      "bar123-Baz",
					Listeners: []listeners.Listener{{Name: "bar123-Bazlis"}},
				},
			},
		},
		{
			name: "multiple lbs, one invalid",
			lbs: []loadbalancers.LoadBalancer{
				{
					Name:      "foo123-Bar",
					Listeners: []listeners.Listener{{Name: "foolis"}},
				},
				{
					Name:      "123bar",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
				{
					Name:      "bar123-Baz",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
			},
			expected: []loadbalancers.LoadBalancer{
				{
					Name:      "foo123-Bar",
					Listeners: []listeners.Listener{{Name: "foolis"}},
				},
				{
					Name:      "bar123-Baz",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
			},
		},
		{
			name: "multiple lbs, one invalid because of invalid listener",
			lbs: []loadbalancers.LoadBalancer{
				{
					Name:      "foo",
					Listeners: []listeners.Listener{{Name: "foolis"}},
				},
				{
					Name:      "bar",
					Listeners: []listeners.Listener{{Name: "123barlis"}},
				},
				{
					Name:      "bar",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
			},
			expected: []loadbalancers.LoadBalancer{
				{
					Name:      "foo",
					Listeners: []listeners.Listener{{Name: "foolis"}},
				},
				{
					Name:      "bar",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
			},
		},
		{
			name: "multiple invalid lbs",
			lbs: []loadbalancers.LoadBalancer{
				{
					Name:      "foo",
					Listeners: []listeners.Listener{{Name: "foolis"}},
				},
				{
					Name:      "0bar",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
				{
					Name:      "foo_bar",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
				{
					Name:      "foobar!",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
				{
					Name:      "foo-bar!",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
				{
					Name:      "-foo",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
				{
					Name:      "foo-",
					Listeners: []listeners.Listener{{Name: "barlis"}},
				},
			},
			expected: []loadbalancers.LoadBalancer{
				{
					Name:      "foo",
					Listeners: []listeners.Listener{{Name: "foolis"}},
				},
			},
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
