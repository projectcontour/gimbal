/*
Copyright 2019 the Gimbal contributors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/projectcontour/gimbal/pkg/buildinfo"
	"github.com/projectcontour/gimbal/pkg/k8s"
	localmetrics "github.com/projectcontour/gimbal/pkg/metrics"
	"github.com/projectcontour/gimbal/pkg/signals"
	"github.com/projectcontour/gimbal/pkg/util"
	"github.com/projectcontour/gimbal/pkg/vmware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	_ "k8s.io/api/core/v1"
)

var (
	printVersion          bool
	gimbalKubeCfgFile     string
	numProcessThreads     int
	backendName           string
	resyncInterval        time.Duration
	debug                 bool
	prometheusListenPort  int
	discovererMetrics     localmetrics.DiscovererMetrics
	gimbalKubeClientQPS   float64
	gimbalKubeClientBurst int
	reconciliationPeriod  time.Duration
	vsphereUrl            string
	vsphereUsername       string
	vspherePassword       string
	vsphereInsecure       bool
)

const (
	envURL      = "VMWARE_URL"
	envUserName = "VMWARE_USERNAME"
	envPassword = "VMWARE_PASSWORD"
	envInsecure = "VMWARE-INSECURE"
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "Show version and quit")
	flag.IntVar(&numProcessThreads, "num-threads", 2, "Specify number of threads to use when processing queue items.")
	flag.StringVar(&gimbalKubeCfgFile, "gimbal-kubecfg-file", "", "Location of kubecfg file for access to gimbal system kubernetes api, defaults to service account tokens")
	flag.StringVar(&backendName, "backend-name", "", "Name of backend (must be unique)")
	flag.DurationVar(&resyncInterval, "resync-interval", time.Minute*30, "Default resync period for watcher to refresh")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging.")
	flag.IntVar(&prometheusListenPort, "prometheus-listen-address", 8080, "The address to listen on for Prometheus HTTP requests")
	flag.Float64Var(&gimbalKubeClientQPS, "gimbal-client-qps", 5, "The maximum queries per second (QPS) that can be performed on the Gimbal Kubernetes API server")
	flag.IntVar(&gimbalKubeClientBurst, "gimbal-client-burst", 10, "The maximum number of queries that can be performed on the Gimbal Kubernetes API server during a burst")
	flag.StringVar(&vsphereUrl, strings.ToLower(envURL), GetEnvString(envURL, "https://username:password@host/sdk"), "vSphere URL.")
	//flag.StringVar(&vsphereUsername, strings.ToLower(envUserName), "", "vSphere username.")
	//flag.StringVar(&vspherePassword, strings.ToLower(envPassword), "", "vSphere password.")
	flag.BoolVar(&vsphereInsecure, strings.ToLower(envInsecure), false, "Verify the server's certificate chain")
	flag.DurationVar(&reconciliationPeriod, "reconciliation-period", 30*time.Second, "The interval of time between reconciliation loop runs.")
	flag.Parse()
}

func main() {
	var log = logrus.New()
	log.Formatter = util.GetFormatter()

	if printVersion {
		fmt.Println("vmware-discoverer")
		fmt.Printf("Version: %s\n", buildinfo.Version)
		fmt.Printf("Git commit: %s\n", buildinfo.GitSHA)
		fmt.Printf("Git tree state: %s\n", buildinfo.GitTreeState)
		os.Exit(0)
	}

	log.Info("Gimbal VMware Discoverer Starting up...")
	log.Infof("Version: %s", buildinfo.Version)
	log.Infof("Backend name: %s", backendName)
	log.Infof("Number of queue worker threads: %d", numProcessThreads)
	log.Infof("Resync interval: %v", resyncInterval)
	log.Infof("Gimbal kubernetes client QPS: %v", gimbalKubeClientQPS)
	log.Infof("Gimbal kubernetes client burst: %d", gimbalKubeClientBurst)
	log.Infof("Reconciliation period: %v", reconciliationPeriod)

	// Init prometheus metrics
	discovererMetrics = localmetrics.NewMetrics("vmware", backendName)
	discovererMetrics.RegisterPrometheus(true)

	// Log info metric
	discovererMetrics.DiscovererInfoMetric(buildinfo.Version)

	if debug {
		log.Level = logrus.DebugLevel
	}

	// Validate cluster name present
	if backendName == "" {
		log.Fatalf("The VMware cluster name must be provided using the `--backend-name` flag")
	}
	// Verify cluster name is passed
	if util.IsInvalidBackendName(backendName) {
		log.Fatalf("The VMware cluster name must be provided using the `--backend-name` flag or the one passed is invalid")
	}
	log.Infof("BackendName is: %s", backendName)

	// Init
	gimbalKubeClient, err := k8s.NewClientWithQPS(gimbalKubeCfgFile, log, float32(gimbalKubeClientQPS), gimbalKubeClientBurst)
	if err != nil {
		log.Fatal("Could not init k8sclient! ", err)
	}

	// Parse URL from string
	u, err := url.Parse(vsphereUrl)
	if err != nil {
		log.Fatal(err)
	}

	// Override username and/or password as required
	processOverride(u)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println()

	// Connect and log in to ESX or vCenter
	c, err := govmomi.NewClient(ctx, u, vsphereInsecure)
	if err != nil {
		log.Fatal(err)
	}

	f := find.NewFinder(c.Client, true)

	// Find one and only datacenter
	dc, err := f.DefaultDatacenter(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)

	pass, _ := u.User.Password()
	tagClient, err := vmware.NewTaggingClient(ctx, log, vsphereUrl, u.User.Username(), pass, vsphereInsecure)
	if err != nil {
		log.Fatal(err)
	}

	reconciler := vmware.NewReconciler(
		backendName,
		gimbalKubeClient,
		reconciliationPeriod,
		c,
		tagClient,
		log,
		numProcessThreads,
		discovererMetrics,
	)

	// set up signals so we handle the first shutdown signal gracefully
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

	// Kick it off
	log.Info("Starting reconciler")
	go reconciler.Run(stopCh)

	<-stopCh
	log.Info("Stopped VMware discoverer")
}

func processOverride(u *url.URL) {
	envUsername := os.Getenv(envUserName)
	envPassword := os.Getenv(envPassword)

	// Override username if provided
	if envUsername != "" {
		var password string
		var ok bool

		if u.User != nil {
			password, ok = u.User.Password()
		}

		if ok {
			u.User = url.UserPassword(envUsername, password)
		} else {
			u.User = url.User(envUsername)
		}
	}

	// Override password if provided
	if envPassword != "" {
		var username string

		if u.User != nil {
			username = u.User.Username()
		}

		u.User = url.UserPassword(username, envPassword)
	}
}

// GetEnvString returns string from environment variable.
func GetEnvString(v string, def string) string {
	r := os.Getenv(v)
	if r == "" {
		return def
	}

	return r
}
