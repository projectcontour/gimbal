package gimbalbench

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestNumberOfBackendEndpoints test the effect of increasing the number of backend endpoints handled by Gimbal
func TestNumberOfBackendEndpoints(fw Framework, numBackendEndpoints []int, requestRate int) error {
	testName := "gimbalbench-backend-endpoints"
	cleanup, err := createTestNamespaces(testName, fw.LoadGenClient, fw.GimbalClient, fw.BackendClient)
	if err != nil {
		cleanup()
		return err
	}
	defer cleanup()

	// Create deployment in backend cluster with zero instances
	log.Printf("Creating deployment in backend cluster")
	dep, svc, err := createNginxDeployment(fw.BackendClient, testName, "nginx", 0)
	if err != nil {
		return err
	}

	// Create ingress for new backend service
	log.Printf("Creating ingress in gimbal cluster")
	discoveredSvc, err := fw.GimbalClient.Core().Services(testName).List(meta_v1.ListOptions{LabelSelector: fmt.Sprintf("gimbal.heptio.com/service=%s", svc.Name)})
	if len(discoveredSvc.Items) != 1 {
		return fmt.Errorf("expected 1 svc, but found %d services with label gimbal.heptio.com/service=%s", len(discoveredSvc.Items), svc.Name)
	}

	host := testName + ".com"
	err = createIngressRoute(fw.ContourCRDClient, testName, testName, discoveredSvc.Items[0], host)
	if err != nil {
		return err
	}

	for _, endpointCount := range numBackendEndpoints {
		// create all the services we need
		log.Printf("Running tests with %d backend services", endpointCount)
		start := time.Now()
		log.Printf("Scaling deployment to %d replicas", endpointCount)

		dep, err = fw.BackendClient.Apps().Deployments(testName).Get(dep.Name, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		ec := int32(endpointCount)
		dep.Spec.Replicas = &ec
		_, err = fw.BackendClient.Apps().Deployments(testName).Update(dep)
		if err != nil {
			return err
		}

		log.Printf("waiting until all %d endpoints have been discovered", endpointCount)
		start = time.Now()
		tick := 5 * time.Second
		timeout := 5 * time.Minute
		waitUntilTrue(tick, timeout, func() (bool, error) {
			endpoints, err := fw.GimbalClient.Core().Endpoints(testName).List(meta_v1.ListOptions{LabelSelector: "gimbal.heptio.com/service=" + svc.Name})
			if err != nil {
				return false, err
			}
			if len(endpoints.Items) != 1 {
				return false, nil
			}
			var total int
			for _, es := range endpoints.Items[0].Subsets {
				total += len(es.Addresses)
			}
			log.Printf("endpoints discovered: %d", total)
			return total == endpointCount, nil
		})

		log.Printf("all %d endpoints have been discovered", endpointCount)
		log.Printf("discovery took %v", time.Now().Sub(start))

		log.Printf("allow backend workloads to warm up (30 secs)")
		time.Sleep(30 * time.Second)

		// run wrk2 against gimbal
		log.Printf("Creating wrk2 job on loadgen cluster")
		jobName := "wrk2-test-num-backend-endpoints-" + strconv.Itoa(endpointCount)
		numConn := 100
		requestRate := 1000
		job := wrkJob(testName, jobName, &fw.Wrk2NodeCount, numConn, requestRate, host, fw.GimbalURL, fw.WrkHostNetwork)

		log.Printf("creating concurrent connections testing job %q", resourceName(job.ObjectMeta))
		log.Printf("concurrent connections: %d", numConn)
		log.Printf("requests per second: %d", requestRate)
		log.Printf("total backend endpoints: %d", endpointCount)
		_, err = fw.LoadGenClient.Batch().Jobs(testName).Create(job)
		if err != nil {
			return err
		}
		job, err = waitForJob(job.Name, testName, fw.LoadGenClient, 120*time.Second)
		if err != nil {
			return err
		}
		err = downloadJobLogs(fw.LoadGenClient, job, filepath.Join(fw.LogsDir, "test-num-endpoints"))
		if err != nil {
			return err
		}
	}

	return nil
}
