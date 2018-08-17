package gimbalbench

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestNumberOfBackendServices tests the effect of increasing the number of backend services that are handled by Gimbal
func TestNumberOfBackendServices(fw Framework, numBackendServices []int, requestRate int) error {
	// Create namespace in all clusters
	testName := fmt.Sprintf("gimbalbench-num-backends-%d", time.Now().Nanosecond())
	cleanup, err := createTestNamespaces(testName, fw.LoadGenClient, fw.GimbalClient, fw.BackendClient)
	if err != nil {
		cleanup()
		return err
	}
	defer cleanup()

	// Create deployment in backend cluster
	log.Printf("Creating deployment in backend cluster")
	dep, svc, err := createNginxDeployment(fw.BackendClient, testName, "nginx", fw.NginxNodeCount)
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
	err = createIngressRoute(fw.ContourCRDClient, testName, "nginx", discoveredSvc.Items[0], host)
	if err != nil {
		return err
	}

	// TODO(abrand): we already have one base service that we created to create the ingress...
	var existingServices = 0
	for _, serviceCount := range numBackendServices {
		// create all the services we need
		log.Printf("Running tests with %d backend services", serviceCount)
		missing := serviceCount - existingServices
		start := time.Now()
		log.Printf("Have %d services. Creating %d extra services", existingServices, missing)
		for i := 0; i < missing; i++ {
			svc := &v1.Service{
				Spec: v1.ServiceSpec{
					Selector: map[string]string{"app": dep.Name},
					Ports:    []v1.ServicePort{{Protocol: v1.ProtocolTCP, Port: 80}},
				},
			}
			svc.Name = fmt.Sprintf("gimbal-test-num-backends-%d", existingServices)
			svc.Namespace = testName
			svc.Labels = map[string]string{"test.gimbal.heptio.com/testName": testName}
			err = retry(5, 1*time.Second, func() error {
				_, err = fw.BackendClient.Core().Services(testName).Create(svc)
				return err
			})
			if err != nil {
				return err
			}
			existingServices++
		}

		log.Printf("creating %d services took %v", missing, time.Now().Sub(start))

		// TODO(abrand): Wait until all the services have been discovered
		log.Printf("waiting until all %d services have been discovered", serviceCount)
		start = time.Now()
		tick := 5 * time.Second
		timeout := 10 * time.Minute
		waitUntilTrue(tick, timeout, func() (bool, error) {
			svcs, err := fw.GimbalClient.Core().Services(testName).List(meta_v1.ListOptions{LabelSelector: "test.gimbal.heptio.com/testName=" + testName})
			if err != nil {
				return false, err
			}
			log.Printf("total services discovered: %d", len(svcs.Items))
			return len(svcs.Items) == serviceCount, nil
		})
		log.Printf("all %d services have been discovered", serviceCount)
		log.Printf("discovery took %v", time.Now().Sub(start))

		// Creating the services hammers kube-proxy. Wait a bit until the dust settles so that nginx has cpus available.
		// We also need to wait a bit until the sockets on the wrk2 side are ready to be used again.
		log.Printf("waiting 90 seconds before running wrk2 job")
		time.Sleep(90 * time.Second)

		// run wrk2 against gimbal
		log.Printf("Creating wrk2 job on loadgen cluster")
		jobName := "wrk2-test-num-backends-" + strconv.Itoa(serviceCount)
		numConn := 100
		requestRate := 1000
		job := wrkJob(testName, jobName, &fw.Wrk2NodeCount, numConn, requestRate, host, fw.GimbalURL, fw.WrkHostNetwork)

		log.Printf("creating concurrent connections testing job %q", resourceName(job.ObjectMeta))
		log.Printf("concurrent connections: %d", numConn)
		log.Printf("requests per second: %d", requestRate)
		log.Printf("total backend services: %d", serviceCount)

		jobClient := fw.LoadGenClient.Batch().Jobs(testName)
		_, err = jobClient.Create(job)
		if err != nil {
			return err
		}

		job, err := waitForJob(job.Name, testName, fw.LoadGenClient, 120*time.Second)
		if err != nil {
			return err
		}
		err = downloadJobLogs(fw.LoadGenClient, job, filepath.Join(fw.LogsDir, "test-num-services"))
		if err != nil {
			return err
		}
	}
	return nil
}
