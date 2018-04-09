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

// +build openstack

package openstack

import (
	"fmt"
	"os"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

func TestListLoadBalancers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping openstack client integration tests")
	}

	c := lbaasClientFromEnv(t)
	projectID := envOrFail(t, "OS_PROJECT_ID")
	lbs, err := c.ListLoadBalancers(projectID)
	if err != nil {
		t.Errorf("failed to list load balancers: %v", err)
	}
	fmt.Println("Load Balancers:", lbs)
	for _, lb := range lbs {
		fmt.Println("Listeners:", lb.Listeners)
	}
}

func TestListPools(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping openstack client integration tests")
	}

	c := lbaasClientFromEnv(t)
	projectID := envOrFail(t, "OS_PROJECT_ID")
	pools, err := c.ListPools(projectID)
	if err != nil {
		t.Errorf("failed to list pools: %v", err)
	}
	fmt.Println("Pools:", pools)
	for _, p := range pools {
		fmt.Println("Pool members:", p.Members)
	}
}

func lbaasClientFromEnv(t *testing.T) LoadBalancerV2Client {
	osAuthOptions := gophercloud.AuthOptions{
		IdentityEndpoint: envOrFail(t, "OS_AUTH_URL"),
		Username:         envOrFail(t, "OS_USERNAME"),
		Password:         envOrFail(t, "OS_PASSWORD"),
		TenantName:       envOrFail(t, "OS_TENANT_NAME"),
		DomainName:       "Default",
	}
	osClient, err := openstack.AuthenticatedClient(osAuthOptions)
	if err != nil {
		t.Fatalf("failed to get openstack client: %v", err)
	}
	lbv2, err := NewLoadBalancerV2(osClient)
	if err != nil {
		t.Fatalf("failed to get LBaaS v2 client: %v", err)
	}
	return *lbv2
}

func envOrFail(t *testing.T, key string) string {
	e := os.Getenv(key)
	if e == "" {
		t.Fatalf("The %q env var must be set", key)
	}
	return e
}
