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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gophercloud/gophercloud"
	gopheropenstack "github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/pools"
	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
)

type OpenstackAuth struct {
	*gophercloud.ProviderClient
	gophercloud.AuthOptions
	log *logrus.Logger
}

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

// NewOpenstackAuth returns an Openstack Auth Client
func NewOpenstackAuth(identityEndpoint, backendName, openstackCertificateAuthorityFile, username, password,
	userDomainName, tenantName string, discovererMetrics *localmetrics.DiscovererMetrics,
	httpClientTimeout time.Duration, log *logrus.Logger) OpenstackAuth {

	// Create and configure client
	osClient, err := gopheropenstack.NewClient(identityEndpoint)
	if err != nil {
		log.Fatalf("Failed to create OpenStack client: %v", err)
	}

	transport := &LogRoundTripper{
		RoundTripper: http.DefaultTransport,
		Log:          log,
		BackendName:  backendName,
		Metrics:      discovererMetrics,
	}

	if openstackCertificateAuthorityFile != "" {
		transport.RoundTripper = httpTransportWithCA(log, openstackCertificateAuthorityFile)
	}

	osClient.HTTPClient = http.Client{
		Transport: transport,
		Timeout:   httpClientTimeout,
	}

	osAuthOptions := gophercloud.AuthOptions{
		IdentityEndpoint: identityEndpoint,
		Username:         username,
		Password:         password,
		DomainName:       userDomainName,
		TenantName:       tenantName,
	}

	return OpenstackAuth{
		ProviderClient: osClient,
		AuthOptions:    osAuthOptions,
		log:            log,
	}
}

// Authenticate authenticates with an Openstack Cluster
func (o *OpenstackAuth) Authenticate() {
	if err := gopheropenstack.Authenticate(o.ProviderClient, o.AuthOptions); err != nil {
		log.Fatalf("Failed to authenticate with OpenStack: %v", err)
	}
	o.log.Info("Success Authenticating with Openstack!")
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

func httpTransportWithCA(log *logrus.Logger, caFile string) http.RoundTripper {
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("Error reading certificate authority for OpenStack: %v", err)
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("Failed to add certificate authority to CA pool. Verify certificate is a valid, PEM-encoded certificate.")
	}
	// Use default transport with CA
	// TODO(abrand): Is there a better way to do this?
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	}
}
