package gimbalbench

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestConcurrentConnections test the effect of increasing the number of concurrent connections handled by Gimbal.
func TestConcurrentConnections(fw Framework, connections []int, requestRate int) error {
	// Setup
	testName := fmt.Sprintf("gimbalbench-concurrent-connections-%d", time.Now().Unix())
	cleanup, err := createTestNamespaces(testName, fw.LoadGenClient, fw.GimbalClient, fw.BackendClient)
	if err != nil {
		cleanup()
		return err
	}
	defer cleanup()

	// Create deployment
	log.Printf("creating deployment in backend cluster")
	_, svc, err := createNginxDeployment(fw.BackendClient, testName, "nginx", fw.NginxNodeCount)
	if err != nil {
		cleanup()
		return err
	}

	// TODO:(abrand) wait until all deployment instances are up and running?
	log.Printf("waiting until service is discovered by gimbal")
	tick := 5 * time.Second
	timeout := 60 * time.Second
	waitUntilTrue(tick, timeout, func() (bool, error) {
		svcs, err := fw.GimbalClient.Core().Services(testName).List(meta_v1.ListOptions{LabelSelector: fmt.Sprintf("gimbal.heptio.com/service=%s", svc.Name)})
		if err != nil {
			return false, err
		}
		return len(svcs.Items) == 1, nil
	})
	log.Printf("service discovered successfully")

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

	// Run tests
	jobClient := fw.LoadGenClient.Batch().Jobs(testName)
	for _, c := range connections {
		// evenly split the number of connections across nodes
		numConn := c / int(fw.Wrk2NodeCount)
		name := "wrk2-connections-" + strconv.Itoa(c)
		job := wrkJob(testName, name, &fw.Wrk2NodeCount, numConn, requestRate, host, fw.GimbalURL, fw.WrkHostNetwork)

		log.Printf("creating concurrent connections testing job %q", resourceName(job.ObjectMeta))
		log.Printf("concurrent connections: %d", c)
		log.Printf("concurrent connections per node: %d", numConn)
		log.Printf("requests per second: %d", requestRate)

		_, err = jobClient.Create(job)
		if err != nil {
			return err
		}
		job, err := waitForJob(job.Name, testName, fw.LoadGenClient, 120*time.Second)
		if err != nil {
			return err
		}
		err = downloadJobLogs(fw.LoadGenClient, job, filepath.Join(fw.LogsDir, "test-concurrent-connections"))
		if err != nil {
			return err
		}
	}
	return nil
}
