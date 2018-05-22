package gimbalbench

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNumberOfIngressRoutes(fw Framework, numIngress []int, requestRate int) error {
	log.Print("Testing the effect of increasing the number of IngressRoute resources Gimbal has to handle.")
	testName := fmt.Sprintf("gimbalbench-num-ingressroutes-%d", time.Now().Nanosecond())
	cleanup, err := createTestNamespaces(testName, fw.LoadGenClient, fw.GimbalClient, fw.BackendClient)
	if err != nil {
		cleanup()
		return err
	}
	defer cleanup()

	_, svc, err := createNginxDeployment(fw.BackendClient, testName, "nginx", fw.NginxNodeCount)
	if err != nil {
		return err
	}

	// Create ingress for new backend service
	discoveredSvc, err := fw.GimbalClient.Core().Services(testName).List(meta_v1.ListOptions{LabelSelector: fmt.Sprintf("gimbal.heptio.com/service=%s", svc.Name)})
	if len(discoveredSvc.Items) != 1 {
		return fmt.Errorf("expected 1 svc, but found %d services with label gimbal.heptio.com/service=%s", len(discoveredSvc.Items), svc.Name)
	}

	var i int
	for _, desiredIngressCount := range numIngress {
		// create ingresses
		log.Printf("Number of desired ingress routes: %d", desiredIngressCount)
		start := time.Now()
		for {
			ingName := fmt.Sprintf("nginx-%d", i)
			err := createIngressRoute(fw.ContourCRDClient, testName, ingName, discoveredSvc.Items[0], fmt.Sprintf("%s-%d.com", testName, i))
			if err != nil {
				return err
			}
			i++
			if i == desiredIngressCount {
				break
			}
		}
		log.Printf("creating ingress routes took %v", time.Now().Sub(start))

		// run wrk2 job
		log.Printf("running wrk2 with a total of %d ingress routes in gimbal", desiredIngressCount)
		jobName := fmt.Sprintf("wrk2-test-num-ingressroutes-%d", desiredIngressCount)
		job := wrkJob(testName, jobName, &fw.Wrk2NodeCount, 100, requestRate, testName+"-0.com", fw.GimbalURL, fw.WrkHostNetwork)
		log.Printf("total ingress routes: %d", desiredIngressCount)
		_, err = fw.LoadGenClient.Batch().Jobs(testName).Create(job)
		if err != nil {
			return err
		}
		job, err = waitForJob(job.Name, testName, fw.LoadGenClient, 120*time.Second)
		if err != nil {
			return err
		}
		err = downloadJobLogs(fw.LoadGenClient, job, filepath.Join(fw.LogsDir, "test-num-ingressroutes"))
		if err != nil {
			return err
		}
	}
	return nil
}
