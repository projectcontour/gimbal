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
	"strings"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

func TestListProjects(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping openstack client integration tests")
	}

	c := identityClientFromEnv(t)
	projects, err := c.ListProjects()
	if err != nil {
		t.Errorf("failed to list projects: %v", err)
	}
	openstackProjectWatchlist := envOrFail(t, "OS_PROJECT_WHITELIST")

	if openstackProjectWatchlist != "" && len(projects) > 0 {
		watchedProjects := strings.Split(openstackProjectWatchlist, ",")
		for _, project := range projects {
			for _, watchedProject := range watchedProjects {
				if watchedProject == project.Name {
					tmp = append(tmp, project)
				}
			}
		}
		projects = tmp
	}

	fmt.Println("Projects:", projects)
}

func identityClientFromEnv(t *testing.T) IdentityV3Client {
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

	identity, err := openstack.NewIdentityV3(osClient)
	if err != nil {
		t.Errorf("Failed to create Identity V3 API client: %v", err)
	}

	return *identity
}

func envOrFail(t *testing.T, key string) string {
	e := os.Getenv(key)
	if e == "" {
		t.Fatalf("The %q env var must be set", key)
	}
	return e
}
