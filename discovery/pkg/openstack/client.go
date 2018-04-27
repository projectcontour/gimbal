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

package openstack

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	gopheropenstack "github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/pools"
)

// IdentityV3Client is a client of the OpenStack Keystone v3 API
type IdentityV3Client struct {
	client *gophercloud.ServiceClient
}

// NewIdentityV3 returns a client of the Keystone v3 API
func NewIdentityV3(provider *gophercloud.ProviderClient) (*IdentityV3Client, error) {
	c, err := gopheropenstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, err
	}
	return &IdentityV3Client{c}, nil
}

// ListProjects returns the list of projects that are available to the user
func (c *IdentityV3Client) ListProjects() ([]projects.Project, error) {
	page, err := projects.List(c.client, projects.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %v", err)
	}
	return projects.ExtractProjects(page)
}

// LoadBalancerV2Client is a client of the OpenStack LBaaS v2 API
type LoadBalancerV2Client struct {
	client *gophercloud.ServiceClient
}

// NewLoadBalancerV2 returns a client of the Load Balancer as a Service v2 API
func NewLoadBalancerV2(provider *gophercloud.ProviderClient) (*LoadBalancerV2Client, error) {
	net, err := gopheropenstack.NewNetworkV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, err
	}
	return &LoadBalancerV2Client{net}, nil
}

// ListLoadBalancers returns the load balancers that exist in the given project
func (c LoadBalancerV2Client) ListLoadBalancers(projectID string) ([]loadbalancers.LoadBalancer, error) {
	lbPage, err := loadbalancers.List(c.client, loadbalancers.ListOpts{TenantID: projectID}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list load balancers: %v", err)
	}

	lbs, err := loadbalancers.ExtractLoadBalancers(lbPage)
	if err != nil {
		return nil, fmt.Errorf("failed to extract load balancers: %v", err)
	}

	lisPage, err := listeners.List(c.client, listeners.ListOpts{TenantID: projectID}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list load balancer listeners: %v", err)
	}

	lis, err := listeners.ExtractListeners(lisPage)
	if err != nil {
		return nil, fmt.Errorf("failed to extract load balancer listeners: %v", err)
	}

	// hydrate each load balancer resource with its listeners
	for i := range lbs {
		lb := &lbs[i]
		var listeners []listeners.Listener
		for _, l := range lis {
			for _, id := range l.Loadbalancers {
				if id.ID == lb.ID {
					listeners = append(listeners, l)
				}
			}
		}
		lb.Listeners = listeners
	}
	return lbs, nil
}

// ListPools returns all load balancer pools that exist in the given project
func (c LoadBalancerV2Client) ListPools(projectID string) ([]pools.Pool, error) {
	page, err := pools.List(c.client, pools.ListOpts{TenantID: projectID}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list listener pools: %v", err)
	}

	ps, err := pools.ExtractPools(page)
	if err != nil {
		return nil, fmt.Errorf("failed extract listener pools: %v", err)
	}

	// add members to each pool
	for i := range ps {
		pool := &ps[i]
		page, err = pools.ListMembers(c.client, pool.ID, pools.ListMembersOpts{TenantID: projectID}).AllPages()
		if err != nil {
			return nil, fmt.Errorf("failed to list members of pool ID %q: %v", pool.ID, err)
		}
		m, err := pools.ExtractMembers(page)
		if err != nil {
			return nil, fmt.Errorf("failed to extract members of pool ID %q: %v", pool.ID, err)
		}
		pool.Members = m
	}

	return ps, nil
}
