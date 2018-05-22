package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	contourv1beta1 "github.com/heptio/contour/apis/contour/v1beta1"
	"github.com/heptio/gimbal/test/performance/internal/gimbalbench"
	"github.com/heptio/gimbal/test/performance/internal/report"
)

var (
	backendKubecfgFile                string
	gimbalKubecfgFile                 string
	loadGenKubecfgFile                string
	gimbalURL                         string
	concurrentConnections             string
	wrk2RequestRate                   int
	wrkHostNetwork                    bool
	backendServicesTest               string
	backendEndpointsTest              string
	gimbalIngressesTest               string
	gimbalIngressRoutesTest           string
	logsDir                           string
	wrk2NodeCount                     int32
	nginxNodeCount                    int32
	gimbalKubernetesDiscoveryTimeTest string
	cleanupKubeconfigs                []string
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := kingpin.New("gimbalbench", "Performance benchmarking tool for Heptio Gimbal.")
	run := app.Command("run", "Run the tests against an existing set of clusters")
	run.Flag("gimbal-url", "The Gimbal LB endpoint for HTTP requests.").Required().StringVar(&gimbalURL)
	run.Flag("backend-kubecfg-file", "Location of the kubeconfig file for the backend cluster.").Required().StringVar(&backendKubecfgFile)
	run.Flag("gimbal-kubecfg-file", "Location of kubeconfig file for the Gimbal cluster.").Required().StringVar(&gimbalKubecfgFile)
	run.Flag("loadgen-kubecfg-file", "Location of kubeconfig file for the LoadGen cluster.").Required().StringVar(&loadGenKubecfgFile)
	run.Flag("rate", "Request rate for wrk2 (requests per second)").Default("1000").IntVar(&wrk2RequestRate)
	run.Flag("host-network", "Whether to put wrk2 in the host network").Default("false").BoolVar(&wrkHostNetwork)
	// flags that provide test paramters start with test-
	run.Flag("test-concurrent-connections", "Comma-separated list of integers for the concurrent connection test.").StringVar(&concurrentConnections)
	run.Flag("test-backend-services", "Comma-separated list of integers for the backend services test.").StringVar(&backendServicesTest)
	run.Flag("test-backend-endpoints", "Comma-separated list of integers for the backend endpoints test.").StringVar(&backendEndpointsTest)
	run.Flag("test-gimbal-ingresses", "Comma-separated list of integers for the number of ingresses test.").StringVar(&gimbalIngressesTest)
	run.Flag("test-gimbal-ingressroutes", "Comma-separated list of integers for the number of ingress routes test.").StringVar(&gimbalIngressRoutesTest)
	run.Flag("test-kubernetes-discovery-time", "Comma-separated list of integers for the number of services to create for the test.").StringVar(&gimbalKubernetesDiscoveryTimeTest)

	reportCmd := app.Command("report", "Generate a report from existing logs files for a previous test run")
	reportCmd.Arg("logs-directory", "The path to the logs directory that contains the logs for a gimbalbench test run").Required().StringVar(&logsDir)

	cleanupCmd := app.Command("clean", "Clean removes all gimbalbench namespaces from the given clusters")
	cleanupCmd.Arg("kubeconfigs", "One or more kubeconfig files for the clusters that should be cleaned").Required().StringsVar(&cleanupKubeconfigs)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case run.FullCommand():
		runTests()
	case reportCmd.FullCommand():
		rep, err := report.Build(logsDir)
		exitOnErr(err)

		err = report.PrettyPrint(os.Stdout, *rep)
		exitOnErr(err)
	case cleanupCmd.FullCommand():
		runClean()
	}
}

func runTests() {
	logsDir = filepath.Join("logs", time.Now().Format("2006-01-02-15-04-05"))
	backendConfig, err := clientcmd.BuildConfigFromFlags("", backendKubecfgFile)
	backendConfig.QPS = 100
	backendConfig.Burst = 500
	exitOnErr(err)
	backendClient, err := kubernetes.NewForConfig(backendConfig)
	exitOnErr(err)

	gimbalConfig, err := clientcmd.BuildConfigFromFlags("", gimbalKubecfgFile)
	gimbalConfig.QPS = 100
	gimbalConfig.Burst = 500
	exitOnErr(err)
	gimbalClient, err := kubernetes.NewForConfig(gimbalConfig)
	exitOnErr(err)

	// Hack alert: The contour generated client is hidden behind the internal pkg in the contour repo
	// hacked this together for now...
	contourCRDConfig := gimbalConfig
	contourCRDConfig.GroupVersion = &contourv1beta1.SchemeGroupVersion
	contourCRDConfig.APIPath = "/apis"
	contourCRDConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(runtime.NewScheme())}
	contourCRDClient, err := restclient.UnversionedRESTClientFor(contourCRDConfig)
	exitOnErr(err)

	loadgenConfig, err := clientcmd.BuildConfigFromFlags("", loadGenKubecfgFile)
	exitOnErr(err)
	loadgenClient, err := kubernetes.NewForConfig(loadgenConfig)
	exitOnErr(err)

	nodes, err := loadgenClient.Core().Nodes().List(meta_v1.ListOptions{LabelSelector: "workload=wrk2"})
	exitOnErr(err)
	wrk2NodeCount = int32(len(nodes.Items))

	if wrk2NodeCount < 1 {
		log.Fatal("no nodes were found with label workload=wrk2")
	}

	nodes, err = backendClient.Core().Nodes().List(meta_v1.ListOptions{LabelSelector: "workload=nginx"})
	exitOnErr(err)
	nginxNodeCount = int32(len(nodes.Items))
	if nginxNodeCount < 1 {
		log.Fatal("no nodes were found with label workload=nginx")
	}

	// Store test configuration
	tc := report.TestConfiguration{
		StartTime:      time.Now(),
		WrkRequestRate: wrk2RequestRate,
		WrkHostNetwork: wrkHostNetwork,
		WrkNodeCount:   int(wrk2NodeCount),
		NginxNodeCount: int(nginxNodeCount),
	}

	b, err := json.Marshal(tc)
	exitOnErr(err)
	err = os.MkdirAll(logsDir, 0755)
	exitOnErr(err)
	err = ioutil.WriteFile(filepath.Join(logsDir, "testconfig.json"), b, 0644)
	exitOnErr(err)

	fw := gimbalbench.Framework{
		GimbalURL:        gimbalURL,
		GimbalClient:     gimbalClient,
		BackendClient:    backendClient,
		LoadGenClient:    loadgenClient,
		ContourCRDClient: contourCRDClient,
		LogsDir:          logsDir,
		NginxNodeCount:   nginxNodeCount,
		Wrk2NodeCount:    wrk2NodeCount,
		WrkHostNetwork:   wrkHostNetwork,
	}
	// Run tests
	if concurrentConnections != "" {
		connections, err := toIntSlice(concurrentConnections)
		exitOnErr(err)

		log.Println("Running concurrent connections test")
		log.Printf("Test cases (total connections): %v", connections)
		err = gimbalbench.TestConcurrentConnections(fw, connections, wrk2RequestRate)
		exitOnErr(err)
	}

	if backendServicesTest != "" {
		services, err := toIntSlice(backendServicesTest)
		exitOnErr(err)

		log.Println("Running backend services test")
		log.Printf("Test cases (total connections): %v", services)
		err = gimbalbench.TestNumberOfBackendServices(fw, services, wrk2RequestRate)
		exitOnErr(err)
	}

	if backendEndpointsTest != "" {
		endpoints, err := toIntSlice(backendEndpointsTest)
		exitOnErr(err)

		log.Println("Running backend endpoints test")
		log.Printf("Test cases (total backend endpoints): %v", endpoints)
		err = gimbalbench.TestNumberOfBackendEndpoints(fw, endpoints, wrk2RequestRate)
		exitOnErr(err)
	}

	if gimbalIngressesTest != "" {
		ingresses, err := toIntSlice(gimbalIngressesTest)
		exitOnErr(err)

		log.Println("Running backend ingresses test")
		log.Printf("Test cases (total backend ingresses): %v", ingresses)
		err = gimbalbench.TestNumberOfIngress(fw, ingresses, wrk2RequestRate)
		exitOnErr(err)
	}

	if gimbalIngressRoutesTest != "" {
		ingressRoutes, err := toIntSlice(gimbalIngressRoutesTest)
		exitOnErr(err)

		log.Println("Running backend ingress routes test")
		log.Printf("Test cases (total ingress routes): %v", ingressRoutes)
		err = gimbalbench.TestNumberOfIngressRoutes(fw, ingressRoutes, wrk2RequestRate)
		exitOnErr(err)
	}

	if gimbalKubernetesDiscoveryTimeTest != "" {
		services, err := toIntSlice(gimbalKubernetesDiscoveryTimeTest)
		exitOnErr(err)

		log.Println("Running kubernetes discoverer test: Time to full discovery")
		log.Printf("Test cases (number of services): %v", services)
		err = gimbalbench.TestKubernetesDiscoveryTime(fw, services)
		exitOnErr(err)
	}
}

func runClean() {
	for _, kc := range cleanupKubeconfigs {
		conf, err := clientcmd.BuildConfigFromFlags("", kc)
		exitOnErr(err)
		c, err := kubernetes.NewForConfig(conf)
		exitOnErr(err)
		ns, err := gimbalbench.ListTestNamespaces(c)
		exitOnErr(err)
		if len(ns) == 0 {
			fmt.Printf("%s: No gimbalbench namespaces were found in the cluster.\n\n", kc)
			continue
		}
		fmt.Printf("%s: The following namespaces will be deleted:\n", kc)
		for _, n := range ns {
			fmt.Println("-", n.Name)
		}
		fmt.Println()
		fmt.Printf("%s: Do you want to proceed? [y/N] ", kc)
		ans, err := bufio.NewReader(os.Stdin).ReadString('\n')
		exitOnErr(err)
		fmt.Println()
		if strings.ToLower(ans) != "y\n" {
			continue
		}
		exitOnErr(gimbalbench.DeleteTestNamespaces(c))
	}
}

func exitOnErr(err error) {
	if err != nil {
		log.Fatalf("unrecoverable error: %v", err)
	}
}

func toIntSlice(csvInts string) ([]int, error) {
	strs := strings.Split(csvInts, ",")
	var r []int
	for _, s := range strs {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		r = append(r, i)
	}
	return r, nil
}
