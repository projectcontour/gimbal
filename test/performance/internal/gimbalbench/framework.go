package gimbalbench

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	contourv1beta1 "github.com/heptio/contour/apis/contour/v1beta1"
	apps_v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// Framework for running gimbalbench tests
type Framework struct {
	GimbalURL        string
	GimbalClient     *kubernetes.Clientset
	BackendClient    *kubernetes.Clientset
	LoadGenClient    *kubernetes.Clientset
	ContourCRDClient restclient.Interface
	LogsDir          string
	Wrk2NodeCount    int32
	NginxNodeCount   int32
	WrkHostNetwork   bool
}

func createIngress(client *kubernetes.Clientset, namespace string, name string, backendSvc v1.Service, host string) (*v1beta1.Ingress, error) {
	ing := &v1beta1.Ingress{}
	ing.Name = name
	ing.Namespace = namespace
	ing.Spec = v1beta1.IngressSpec{
		Rules: []v1beta1.IngressRule{{
			Host: host,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{{
						Path: "/",
						Backend: v1beta1.IngressBackend{
							ServiceName: backendSvc.Name,
							ServicePort: intstr.FromInt(int(backendSvc.Spec.Ports[0].Port)),
						},
					}},
				},
			},
		}},
	}
	return client.ExtensionsV1beta1().Ingresses(namespace).Create(ing)
}

func createIngressRoute(client restclient.Interface, namespace string, name string, backendSvc v1.Service, host string) error {
	ing := &contourv1beta1.IngressRoute{}
	ing.APIVersion = "contour.heptio.com/v1beta1"
	ing.Kind = "IngressRoute"
	ing.Name = name
	ing.Namespace = namespace
	ing.Spec = contourv1beta1.IngressRouteSpec{
		VirtualHost: &contourv1beta1.VirtualHost{
			Fqdn: host,
		},
		Routes: []contourv1beta1.Route{
			{
				Match: "/",
				Services: []contourv1beta1.Service{
					{
						Name: backendSvc.Name,
						Port: int(backendSvc.Spec.Ports[0].Port),
					},
				},
			},
		},
	}
	return client.Post().Namespace(namespace).Resource("ingressroutes").Body(ing).Do().Error()
}

func createTestNamespaces(namespace string, loadGenClient, gimbalClient, backendClient *kubernetes.Clientset) (func(), error) {
	log.Printf("Creating namespace %q in all clusters", namespace)
	var cleanup []func()
	cleanupFunc := func() {
		for _, f := range cleanup {
			f()
		}
	}
	ns := &v1.Namespace{}
	ns.Name = namespace
	_, err := loadGenClient.Core().Namespaces().Create(ns)
	if err != nil {
		return cleanupFunc, err
	}
	cleanup = append(cleanup, func() {
		err := loadGenClient.Core().Namespaces().Delete(ns.Name, &meta_v1.DeleteOptions{})
		if err != nil {
			log.Printf("failed to delete namespace %q in load gen cluster. Must be deleted manually.", ns.Name)
			return
		}
		log.Printf("deleted namespace %s in loadgen cluster", ns.Name)
	})
	_, err = backendClient.Core().Namespaces().Create(ns)
	// TODO: hack because we are using the same cluster for both load gen and backend
	if err != nil && !errors.IsAlreadyExists(err) {
		return cleanupFunc, err
	}
	cleanup = append(cleanup, func() {
		err = backendClient.Core().Namespaces().Delete(ns.Name, &meta_v1.DeleteOptions{})
		if err != nil {
			log.Printf("failed to delete namespace %q in backend cluster. Must be deleted manually.", ns.Name)
			return
		}
		log.Printf("deleted namespace %s in backend cluster", ns.Name)
	})
	_, err = gimbalClient.Core().Namespaces().Create(ns)
	if err != nil && !errors.IsAlreadyExists(err) {
		return cleanupFunc, err
	}

	cleanup = append(cleanup, func() {
		err = gimbalClient.Core().Namespaces().Delete(ns.Name, &meta_v1.DeleteOptions{})
		if err != nil {
			log.Printf("failed to delete namespace %q in gimbal cluster. Must be deleted manually.", ns.Name)
			return
		}
		log.Printf("deleted namespace %s in gimbal cluster", ns.Name)
	})
	return cleanupFunc, nil
}

func downloadJobLogs(client *kubernetes.Clientset, job *batchv1.Job, destinationDir string) error {
	// Ensure destination dir exists
	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return fmt.Errorf("error creating log directory: %v", err)
	}
	log.Printf("Grabbing logs from all pods for job %q", resourceName(job.ObjectMeta))
	pods, err := client.Core().Pods(job.Namespace).List(meta_v1.ListOptions{
		LabelSelector: labelSelectorToStringSelector(job.Spec.Selector),
	})
	if err != nil {
		return fmt.Errorf("error listing pods in namespace %q with selector %q", job.Namespace, job.Spec.Selector.String())
	}
	podClient := client.Core().Pods(job.Namespace)
	for _, p := range pods.Items {
		req := podClient.GetLogs(p.Name, &v1.PodLogOptions{})
		logStream, err := req.Stream()
		if err != nil {
			return fmt.Errorf("error streaming logs for pod %q", resourceName(p.ObjectMeta))
		}
		f, err := os.Create(filepath.Join(destinationDir, p.Name+".log"))
		if err != nil {
			return fmt.Errorf("failed to create file for logs: %v", err)
		}
		log.Printf("saving logs of pod %q into local file at %q", resourceName(p.ObjectMeta), f.Name())
		_, err = io.Copy(f, logStream)
		if err != nil {
			return fmt.Errorf("error writing pod %q logs to file %q", resourceName(p.ObjectMeta), f.Name())
		}
		logStream.Close()
		f.Close()
	}
	return nil
}

func waitForJob(name string, namespace string, client *kubernetes.Clientset, deadline time.Duration) (*batchv1.Job, error) {
	log.Printf("waiting until job %s/%s completes", namespace, name)
	tick := time.NewTicker(10 * time.Second)
	done := time.After(deadline)
	for {
		select {
		case <-done:
			return nil, fmt.Errorf(`timed out waiting for job "%s/%s" to complete`, namespace, name)
		case <-tick.C:
			job, err := client.Batch().Jobs(namespace).Get(name, meta_v1.GetOptions{})
			if err != nil {
				log.Printf(`error getting job "%s/%s" (will retry): %v`, namespace, name, err)
				continue
			}
			// TODO: use job condition here?
			if job.Status.CompletionTime != nil {
				return job, nil
			}
			log.Printf(`job "%s/%s" has not completed yet. waiting...`, namespace, name)
		}
	}
}

func createNginxDeployment(client *kubernetes.Clientset, namespace string, name string, replicas int32) (*apps_v1.Deployment, *v1.Service, error) {
	dep, err := client.Apps().Deployments(namespace).Create(nginxDeployment(namespace, name, replicas))
	if err != nil {
		return nil, nil, err
	}
	log.Printf("created nginx deployment with %d replicas", replicas)

	log.Printf("exposing deployment via service")
	svc := &v1.Service{
		Spec: v1.ServiceSpec{
			Selector: dep.Spec.Template.Labels,
			Ports:    []v1.ServicePort{{Protocol: v1.ProtocolTCP, Port: 80}},
		},
	}
	svc.Name = "nginx"
	svc.Namespace = namespace
	_, err = client.Core().Services(namespace).Create(svc)
	if err != nil {
		return nil, nil, err
	}
	return dep, svc, nil
}

func nginxDeployment(ns string, name string, replicas int32) *apps_v1.Deployment {
	dep := &apps_v1.Deployment{}
	dep.Namespace = ns
	dep.Name = name
	labels := map[string]string{"app": name}
	dep.Spec = apps_v1.DeploymentSpec{
		Selector: &meta_v1.LabelSelector{
			MatchLabels: labels,
		},
		Replicas: &replicas,
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "nginx",
						Image: "nginx",
					},
				},
				NodeSelector: map[string]string{"workload": "nginx"},
				Affinity: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 1,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &meta_v1.LabelSelector{MatchExpressions: []meta_v1.LabelSelectorRequirement{{
										Key:      "app",
										Operator: meta_v1.LabelSelectorOpIn,
										Values:   []string{name}}},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
			},
		},
	}
	dep.Spec.Template.Labels = labels
	return dep
}

func wrkJob(ns string, name string, completions *int32, connections int, requestRate int, hostHeader string, gimbalURL string, wrkHostNetwork bool) *batchv1.Job {
	log.Printf("Creating wrk job: connections=%d, requestRate=%d, hostHeader=%q", connections, requestRate, hostHeader)
	job := &batchv1.Job{
		Spec: batchv1.JobSpec{
			Completions: completions,
			Parallelism: completions,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "wrk2",
							Image: "bootjp/wrk2",
							Command: []string{
								"wrk",
								"--latency",
								"--duration", "60",
								"--threads", "20",
								"--connections", strconv.Itoa(connections),
								"--rate", strconv.Itoa(requestRate),
								"--header", fmt.Sprintf("Host: %s", hostHeader),
								gimbalURL,
							},
						},
					},
					HostNetwork:   wrkHostNetwork,
					RestartPolicy: v1.RestartPolicyNever,
					NodeSelector:  map[string]string{"workload": "wrk2"},
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &meta_v1.LabelSelector{MatchExpressions: []meta_v1.LabelSelectorRequirement{{
										Key:      "app",
										Operator: meta_v1.LabelSelectorOpIn,
										Values:   []string{"wrk2"}}},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
			},
		},
	}
	job.Spec.Template.Labels = map[string]string{"workload": "wrk2"}
	job.Namespace = ns
	job.Name = name
	return job
}

func resourceName(meta meta_v1.ObjectMeta) string {
	return meta.Namespace + "/" + meta.Name
}

func labelSelectorToStringSelector(ls *meta_v1.LabelSelector) string {
	var s []string
	for k, v := range ls.MatchLabels {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(s, ",")
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

func exitOnErr(err error) {
	if err != nil {
		log.Fatalf("unrecoverable error: %v", err)
	}
}

func retry(n int, sleep time.Duration, f func() error) error {
	i := 0
	var err error
	for {
		err = f()
		if err == nil {
			return err
		}
		if i == n-1 {
			return err
		}
		log.Printf("retrying operation... error: %v", err)
		time.Sleep(sleep)
		i++
	}
}

func waitUntilTrue(pause time.Duration, timeout time.Duration, f func() (bool, error)) error {
	var done bool
	var err error
	ticker := time.NewTicker(pause)
	timesup := time.After(timeout)
	for !done {
		select {
		case <-timesup:
			return fmt.Errorf("waited %v for condition to be true", timeout)
		case <-ticker.C:
			done, err = f()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
