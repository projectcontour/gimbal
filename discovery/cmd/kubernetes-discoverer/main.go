/*
Copyright 2018 Heptio Inc.
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
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/heptio/gimbal/discovery/pkg/buildinfo"
	"github.com/heptio/gimbal/discovery/pkg/k8s"
	"github.com/heptio/gimbal/discovery/pkg/signals"
	"github.com/heptio/gimbal/discovery/pkg/util"
	"github.com/sirupsen/logrus"
	_ "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
)

var (
	printVersion          bool
	gimbalKubeCfgFile     string
	discovererKubeCfgFile string
	numProcessThreads     int
	workingNamespace      string
	clusterName           string
	resyncInterval        time.Duration
	debug                 bool
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "Show version and quit")
	flag.IntVar(&numProcessThreads, "num-threads", 2, "Specify number of threads to use when processing queue items.")
	flag.StringVar(&gimbalKubeCfgFile, "gimbal-kubecfg-file", "", "Location of kubecfg file for access to gimbal system kubernetes api, defaults to service account tokens")
	flag.StringVar(&discovererKubeCfgFile, "discover-kubecfg-file", "", "Location of kubecfg file for access to remote discover system kubernetes api")
	flag.StringVar(&clusterName, "cluster-name", "", "Name of cluster")
	flag.DurationVar(&resyncInterval, "resync-interval", time.Minute*30, "Default resync period for watcher to refresh")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging.")
	flag.Parse()
}

func main() {
	var log = logrus.New()
	log.Formatter = util.GetFormatter()

	if printVersion {
		fmt.Println("kubernetes-discoverer")
		fmt.Printf("Version: %s\n", buildinfo.Version)
		fmt.Printf("Git commit: %s\n", buildinfo.GitSHA)
		fmt.Printf("Git tree state: %s\n", buildinfo.GitTreeState)
		os.Exit(0)
	}

	if debug {
		log.Level = logrus.DebugLevel
	}

	// Verify cluster name is passed
	if clusterName == "" {
		log.Fatalf("`cluster-name` arg is required!")
	}

	// Discovered cluster is passed.
	if discovererKubeCfgFile == "" {
		log.Fatalf("`discover-kubecfg-file` arg is required!")
	}

	log.Info("Gimbal Discoverer Starting up...")

	// Init
	gimbalKubeClient, err := k8s.NewClient(gimbalKubeCfgFile, log)
	if err != nil {
		log.Error("Could not init k8sclient! ", err)
	}

	k8sDiscovererClient, err := k8s.NewClient(discovererKubeCfgFile, log)
	if err != nil {
		log.Error("Could not init k8s discoverer client! ", err)
	}

	log.Info("Starting shared informer, resync interval is: ", resyncInterval)

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(k8sDiscovererClient, resyncInterval)

	c := k8s.NewController(log, gimbalKubeClient, kubeInformerFactory, clusterName, numProcessThreads)
	if err != nil {
		log.Error("Could not init Controller! ", err)
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	go kubeInformerFactory.Start(stopCh)

	// Kick it off
	if err = c.Run(stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}

}
