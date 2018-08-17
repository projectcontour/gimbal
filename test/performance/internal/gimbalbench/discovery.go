package gimbalbench

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesDiscoveryTestResult summarizes the results of the KubernetesDiscoveryTime test
type KubernetesDiscoveryTestResult struct {
	ServiceCount         int
	TimeToFirstDiscovery time.Duration
	TimeToFullDiscovery  time.Duration
	TimeToDiscoverNew    time.Duration
	TimeToDiscoverUpdate time.Duration
}

// TestKubernetesDiscoveryTime tests the time it takes to discover services
func TestKubernetesDiscoveryTime(framework Framework, numServices []int) error {
	for _, desiredServices := range numServices {
		testName := fmt.Sprintf("gimbalbench-kubernetes-discovery-time-%d-services", desiredServices)
		cleanup, err := createTestNamespaces(testName, framework.LoadGenClient, framework.GimbalClient, framework.BackendClient)
		if err != nil {
			cleanup()
			return err
		}

		// wrap so that we can defer cleanup
		err = func() error {
			defer cleanup()

			// 0. Ensure the k8s discoverer has been deployed.
			// TODO(abrand): name of discoverer is hard coded
			dep, err := framework.GimbalClient.Apps().Deployments("gimbal-discovery").Get("k8s-kubernetes-discoverer", meta_v1.GetOptions{})
			if err != nil {
				return err
			}

			var zero int32
			var one int32 = 1
			dep.Spec.Replicas = &zero
			dep, err = framework.GimbalClient.Apps().Deployments("gimbal-discovery").Update(dep)
			if err != nil {
				return err
			}

			// Wait until no pods are running
			tick := 2 * time.Second
			timeout := 30 * time.Second
			log.Printf("waiting until k8s-kubernetes-discoverer has zero pods running.")
			waitUntilTrue(tick, timeout, func() (bool, error) {
				pods, err := framework.GimbalClient.Core().Pods("gimbal-discovery").List(meta_v1.ListOptions{LabelSelector: "app=kubernetes-discoverer"})
				if err != nil {
					return false, err
				}
				log.Printf("k8s-kubernetes-discoverer running pods: %d", len(pods.Items))
				return len(pods.Items) == 0, nil
			})

			start := time.Now()
			log.Printf("Creating %d services", desiredServices)
			for i := 0; i < desiredServices; i++ {
				svc := &v1.Service{
					Spec: v1.ServiceSpec{
						Selector: map[string]string{"app": dep.Name},
						Ports:    []v1.ServicePort{{Protocol: v1.ProtocolTCP, Port: 80}},
					},
				}
				svc.Name = fmt.Sprintf("gimbal-test-num-backends-%d", i)
				svc.Namespace = testName
				svc.Labels = map[string]string{"test.gimbal.heptio.com/testName": testName}
				err = retry(5, 1*time.Second, func() error {
					_, err = framework.BackendClient.Core().Services(testName).Create(svc)
					return err
				})
				if err != nil {
					return err
				}
			}
			log.Printf("creating %d services took %v", desiredServices, time.Now().Sub(start))

			// 2. Start a clock and start the discoverer
			dep, err = framework.GimbalClient.Apps().Deployments("gimbal-discovery").Get("k8s-kubernetes-discoverer", meta_v1.GetOptions{})
			if err != nil {
				return err
			}
			dep.Spec.Replicas = &one
			dep, err = framework.GimbalClient.Apps().Deployments("gimbal-discovery").Update(dep)
			if err != nil {
				return err
			}

			// Wait until discoverer is up
			tick = 500 * time.Millisecond
			timeout = 10 * time.Second
			log.Printf("waiting until discoverer has one replica running")
			waitUntilTrue(tick, timeout, func() (bool, error) {
				dep, err = framework.GimbalClient.Apps().Deployments("gimbal-discovery").Get("k8s-kubernetes-discoverer", meta_v1.GetOptions{})
				if err != nil {
					return false, err
				}
				log.Printf("discoverer has %d ready replicas", dep.Status.ReadyReplicas)
				return dep.Status.ReadyReplicas == 1, nil
			})

			start = time.Now()

			// 3. Wait until all services have been discovered
			log.Printf("waiting until all %d services have been discovered", desiredServices)
			start = time.Now()
			tick = 1 * time.Second
			timeout = 20 * time.Minute
			var firstDiscovered *time.Duration
			waitUntilTrue(tick, timeout, func() (bool, error) {
				svcs, err := framework.GimbalClient.Core().Services(testName).List(meta_v1.ListOptions{LabelSelector: "test.gimbal.heptio.com/testName=" + testName})
				if err != nil {
					return false, err
				}
				log.Printf("total services discovered: %d", len(svcs.Items))
				if len(svcs.Items) > 0 && firstDiscovered == nil {
					t := time.Now().Sub(start)
					firstDiscovered = &t
				}
				return len(svcs.Items) == desiredServices, nil
			})
			fullDiscovey := time.Now().Sub(start)

			// 4. Add a new service, and measure how long it takes to discover
			log.Println("Adding an additional service to see how long it takes to discover")
			svc := &v1.Service{
				Spec: v1.ServiceSpec{
					Selector: map[string]string{"app": dep.Name},
					Ports:    []v1.ServicePort{{Protocol: v1.ProtocolTCP, Port: 80}},
				},
			}
			svc.Name = "gimbal-test-num-backends-additional"
			svc.Namespace = testName
			svc.Labels = map[string]string{"test.gimbal.heptio.com/testName": testName}
			err = retry(5, 1*time.Second, func() error {
				_, err = framework.BackendClient.Core().Services(testName).Create(svc)
				return err
			})
			if err != nil {
				return err
			}

			// Wait until the service is discovered
			start = time.Now()
			tick = 250 * time.Millisecond
			timeout = 600 * time.Second
			waitUntilTrue(tick, timeout, func() (bool, error) {
				svcs, err := framework.GimbalClient.Core().Services(testName).List(meta_v1.ListOptions{LabelSelector: "gimbal.heptio.com/service=gimbal-test-num-backends-additional"})
				if err != nil {
					return false, err
				}
				if len(svcs.Items) != 1 {
					log.Printf("did not find service %q.. retrying", svc.Name)
					return false, nil
				}
				return true, nil
			})
			timeToDiscoverNew := time.Now().Sub(start)
			log.Println("discovered new service")

			// Update the service and check how long it takes for the update to be discovered
			log.Println("Updating service to see how long it takes to discover the update")
			svc, err = framework.BackendClient.Core().Services(testName).Get(svc.Name, meta_v1.GetOptions{})
			if err != nil {
				return err
			}
			svc.Labels["foo"] = "bar"
			_, err = framework.BackendClient.Core().Services(testName).Update(svc)
			if err != nil {
				return err
			}

			// Wait until the service is updated
			start = time.Now()
			tick = 250 * time.Millisecond
			timeout = 600 * time.Second
			waitUntilTrue(tick, timeout, func() (bool, error) {
				svcs, err := framework.GimbalClient.Core().Services(testName).List(meta_v1.ListOptions{LabelSelector: "gimbal.heptio.com/service=gimbal-test-num-backends-additional"})
				if err != nil {
					return false, err
				}
				if len(svcs.Items) == 1 && svcs.Items[0].Labels["foo"] == "bar" {
					return true, nil
				}
				log.Printf("service has not been updated.. retrying...")
				return false, nil
			})
			timeToDiscoverUpdate := time.Now().Sub(start)
			log.Println("discovered update")

			r := KubernetesDiscoveryTestResult{
				ServiceCount:         desiredServices,
				TimeToFirstDiscovery: *firstDiscovered,
				TimeToFullDiscovery:  fullDiscovey,
				TimeToDiscoverNew:    timeToDiscoverNew,
				TimeToDiscoverUpdate: timeToDiscoverUpdate,
			}

			b, err := json.Marshal(r)
			if err != nil {
				return fmt.Errorf("error marshaling test result into JSON: %v", err)
			}

			resFile := filepath.Join(framework.LogsDir, "test-kubernetes-service-discovery-time", fmt.Sprintf("%d-services.json", desiredServices))
			if err := os.MkdirAll(filepath.Dir(resFile), 0755); err != nil {
				return err
			}
			if err := ioutil.WriteFile(resFile, b, 0644); err != nil {
				return err
			}

			log.Printf("all %d services have been discovered", desiredServices)
			log.Printf("Result:")
			log.Printf("- Number of services: %d", desiredServices)
			log.Printf("- Time to first discovery: %v", firstDiscovered)
			log.Printf("- Time to full discovery: %v", fullDiscovey)
			log.Printf("- Time to discover new: %v", timeToDiscoverNew)
			log.Printf("- Time to discover update: %v", timeToDiscoverUpdate)
			log.Println()
			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}
