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

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gophercloud/gophercloud"
	gopheropenstack "github.com/gophercloud/gophercloud/openstack"
	"github.com/heptio/gimbal/discovery/pkg/buildinfo"
	"github.com/heptio/gimbal/discovery/pkg/k8s"
	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/heptio/gimbal/discovery/pkg/openstack"
	"github.com/heptio/gimbal/discovery/pkg/signals"
	"github.com/heptio/gimbal/discovery/pkg/util"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	printVersion                      bool
	gimbalKubeCfgFile                 string
	discoverStackCfgFile              string
	backendName                       string
	numProcessThreads                 int
	debug                             bool
	reconciliationPeriod              time.Duration
	httpClientTimeout                 time.Duration
	openstackCertificateAuthorityFile string
	prometheusListenPort              int
	discovererMetrics                 localmetrics.DiscovererMetrics
	log                               *logrus.Logger
	gimbalKubeClientQPS               float64
	gimbalKubeClientBurst             int
)

const (
	clusterType           = "openstack"
	defaultUserDomainName = "Default"
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "Show version and quit")
	flag.StringVar(&gimbalKubeCfgFile, "gimbal-kubecfg-file", "", "Location of kubecfg file for access to gimbal system kubernetes api, defaults to service account tokens")
	flag.StringVar(&backendName, "backend-name", "", "Name of cluster (must be unique)")
	flag.IntVar(&numProcessThreads, "num-threads", 2, "Specify number of threads to use when processing queue items.")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging.")
	flag.DurationVar(&reconciliationPeriod, "reconciliation-period", 30*time.Second, "The interval of time between reconciliation loop runs.")
	flag.DurationVar(&httpClientTimeout, "http-client-timeout", 5*time.Second, "The HTTP client request timeout.")
	flag.StringVar(&openstackCertificateAuthorityFile, "openstack-certificate-authority", "", "Path to cert file of the OpenStack API certificate authority.")
	flag.IntVar(&prometheusListenPort, "prometheus-listen-address", 8080, "The address to listen on for Prometheus HTTP requests")
	flag.Float64Var(&gimbalKubeClientQPS, "gimbal-client-qps", 5, "The maximum queries per second (QPS) that can be performed on the Gimbal Kubernetes API server")
	flag.IntVar(&gimbalKubeClientBurst, "gimbal-client-burst", 10, "The maximum number of queries that can be performed on the Gimbal Kubernetes API server during a burst")
	flag.Parse()
}

func main() {
	if printVersion {
		fmt.Println("openstack-discoverer")
		fmt.Printf("Version: %s\n", buildinfo.Version)
		fmt.Printf("Git commit: %s\n", buildinfo.GitSHA)
		fmt.Printf("Git tree state: %s\n", buildinfo.GitTreeState)
		os.Exit(0)
	}

	log = logrus.New()
	log.Formatter = util.GetFormatter()
	if debug {
		log.Level = logrus.DebugLevel
	}

	log.Info("Gimbal OpenStack Discoverer Starting up...")
	log.Infof("Version: %s", buildinfo.Version)
	log.Infof("Backend name: %s", backendName)
	log.Infof("Number of queue worker threads: %d", numProcessThreads)
	log.Infof("Reconciliation period: %v", reconciliationPeriod)
	log.Infof("Gimbal kubernetes client QPS: %v", gimbalKubeClientQPS)
	log.Infof("Gimbal kubernetes client burst: %d", gimbalKubeClientBurst)

	// Init prometheus metrics
	discovererMetrics = localmetrics.NewMetrics("openstack", backendName)
	discovererMetrics.RegisterPrometheus(true)

	// Log info metric
	discovererMetrics.DiscovererInfoMetric(buildinfo.Version)

	// Validate cluster name
	if util.IsInvalidBackendName(backendName) {
		log.Fatalf("The Kubernetes cluster name must be provided using the `--backend-name` flag or the one passed is invalid")
	}
	log.Infof("BackendName is: %s", backendName)

	gimbalKubeClient, err := k8s.NewClientWithQPS(gimbalKubeCfgFile, log, float32(gimbalKubeClientQPS), gimbalKubeClientBurst)
	if err != nil {
		log.Fatal("Failed to create kubernetes client", err)
	}

	username := os.Getenv("OS_USERNAME")
	if username == "" {
		log.Fatal("The OpenStack username must be provided using the OS_USERNAME environment variable.")
	}
	password := os.Getenv("OS_PASSWORD")
	if password == "" {
		log.Fatal("The OpenStack password must be provided using the OS_PASSWORD environment variable.")
	}
	identityEndpoint := os.Getenv("OS_AUTH_URL")
	if identityEndpoint == "" {
		log.Fatal("The OpenStack Authentication URL must be provided using the OS_AUTH_URL environment variable.")
	}
	tenantName := os.Getenv("OS_TENANT_NAME")
	if tenantName == "" {
		log.Fatal("The OpenStack tenant name must be provided using the OS_TENANT_NAME environment variable")
	}
	userDomainName := os.Getenv("OS_USER_DOMAIN_NAME")
	if userDomainName == "" {
		log.Warnf("The OS_USER_DOMAIN_NAME environment variable was not set. Using %q as the OpenStack user domain name.", defaultUserDomainName)
		userDomainName = defaultUserDomainName
	}

	// Create and configure client
	osClient, err := gopheropenstack.NewClient(identityEndpoint)
	if err != nil {
		log.Fatalf("Failed to create OpenStack client: %v", err)
	}

	transport := &openstack.LogRoundTripper{
		RoundTripper: http.DefaultTransport,
		Log:          log,
		BackendName:  backendName,
		Metrics:      &discovererMetrics,
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

	if err := gopheropenstack.Authenticate(osClient, osAuthOptions); err != nil {
		log.Fatalf("Failed to authenticate with OpenStack: %v", err)
	}

	identity, err := openstack.NewIdentityV3(osClient)
	if err != nil {
		log.Fatalf("Failed to create Identity V3 API client: %v", err)
	}

	lbv2, err := openstack.NewLoadBalancerV2(osClient)
	if err != nil {
		log.Fatalf("Failed to create Network V2 API client: %v", err)
	}

	reconciler := openstack.NewReconciler(
		backendName,
		clusterType,
		gimbalKubeClient,
		reconciliationPeriod,
		lbv2,
		identity,
		log,
		numProcessThreads,
		discovererMetrics,
	)
	stopCh := signals.SetupSignalHandler()

	go func() {
		// Expose the registered metrics via HTTP.
		http.Handle("/metrics", promhttp.HandlerFor(discovererMetrics.Registry, promhttp.HandlerOpts{}))
		srv := &http.Server{Addr: fmt.Sprintf(":%d", prometheusListenPort)}
		log.Info("Listening for Prometheus metrics on port: ", prometheusListenPort)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
		<-stopCh
		log.Info("Shutting down Prometheus server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	log.Info("Starting reconciler")
	go reconciler.Run(stopCh)

	<-stopCh
	log.Info("Stopped OpenStack discoverer")
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
