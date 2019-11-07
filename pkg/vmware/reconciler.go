package vmware

import (
	"context"
	"fmt"
	"time"

	localmetrics "github.com/projectcontour/gimbal/pkg/metrics"
	"github.com/projectcontour/gimbal/pkg/sync"
	"github.com/projectcontour/gimbal/pkg/translator"
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type VirtualMachine struct {
	Name        string
	UUID        string
	ProjectName string
	Application string
	Port        string
	IPAddress   string
}

// The Reconciler connects to an VMware cluster and makes sure that VMs
// defined in the cluster are reflected in the Gimbal Kubernetes
// cluster as Services and Endpoint. The Reconciler runs on a configurable
// interval.
type Reconciler struct {
	// BackendName is the name of the OpenStack cluster
	BackendName string
	ClusterType string

	// VMware client to query for resources
	Client        *govmomi.Client
	TaggingClient *TaggingClient

	// GimbalKubeClient is the client of the Kubernetes cluster where Gimbal is running
	GimbalKubeClient kubernetes.Interface

	// Interval between reconciliation loops
	SyncPeriod time.Duration
	Logger     *logrus.Logger
	syncqueue  sync.Queue

	Metrics localmetrics.DiscovererMetrics
}

// NewReconciler returns an VMware reconciler
func NewReconciler(backendName string, gimbalKubeClient kubernetes.Interface, syncPeriod time.Duration, client *govmomi.Client,
	taggingClient *TaggingClient, log *logrus.Logger, queueWorkers int, metrics localmetrics.DiscovererMetrics) Reconciler {

	return Reconciler{
		BackendName:      backendName,
		GimbalKubeClient: gimbalKubeClient,
		SyncPeriod:       syncPeriod,
		Client:           client,
		TaggingClient:    taggingClient,
		Logger:           log,
		Metrics:          metrics,
		syncqueue:        sync.NewQueue(log, gimbalKubeClient, queueWorkers, metrics),
	}
}

// Run starts the reconciler
func (r *Reconciler) Run(stop <-chan struct{}) {
	go r.syncqueue.Run(stop)

	ticker := time.NewTicker(r.SyncPeriod)
	defer ticker.Stop()

	// Perform an initial reconciliation
	r.reconcile()

	// Perform reconciliation on every tick
	for {
		select {
		case <-stop:
			r.Logger.Info("Stopping vmware reconciler")
			return
		case <-ticker.C:
			r.reconcile()
		}
	}
}

func (r *Reconciler) reconcile() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Calculate cycle time
	start := time.Now()

	log := r.Logger
	log.Info("reconciling VMs")

	// Create view of VirtualMachine objects
	m := view.NewManager(r.Client.Client)

	v, err := m.CreateContainerView(ctx, r.Client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		log.Error(err)
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all machines
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "customValue"}, &vms)
	if err != nil {
		log.Error(err)
	}

	var virtualMachines []VirtualMachine
	for _, vm := range vms {

		err = property.DefaultCollector(r.Client.Client).Retrieve(ctx, []types.ManagedObjectReference{vm.ManagedEntity.Reference()}, []string{"customValue"}, &vms)
		if err != nil {
			log.Error("-- err: ", err)
		}

		matches := func(key int32) bool {
			return true
		}

		m, err := object.GetCustomFieldsManager(r.Client.Client)
		if err != nil {
			log.Error(err)
		}

		field, err := m.Field(ctx)
		if err != nil {
			log.Error(err)
		}

		projectName := ""
		application := ""
		port := ""

		for i := range vm.CustomValue {
			val := vm.CustomValue[i].(*types.CustomFieldStringValue)

			if !matches(val.Key) {
				continue
			}
			switch field.ByKey(val.Key).Name {
			case "gimbal-project":
				projectName = val.Value
			case "gimbal-port":
				port = val.Value
			case "gimbal-application":
				application = val.Value
			}
		}

		//// Get a set of all tags for this VirtualMachine
		//tags, err := r.TaggingClient.manager.GetAttachedTags(ctx, vm.ManagedEntity)
		//if err != nil {
		//	log.Error(err)
		//}
		//
		//projectName, application, port := lookupTags(tags)

		if projectName == "" {
			log.Errorf("Missing `gimbal-project` label on VM")
			continue
		}
		if application == "" {
			log.Errorf("Missing `gimbal-application` label on VM")
			continue
		}
		if port == "" {
			log.Errorf("Missing `gimbal-port` label on VM")
			continue
		}

		virtualMachines = append(virtualMachines, VirtualMachine{
			ProjectName: projectName,
			Application: application,
			Port:        port,
			IPAddress:   vm.Summary.Guest.IpAddress,
			UUID:        vm.Summary.Config.Uuid,
			Name:        vm.Summary.Config.Name,
		})
	}

	// Group services by `ProjectName`
	svcs := map[string][]VirtualMachine{}
	for _, v := range virtualMachines {
		svcs[v.ProjectName] = append(svcs[v.ProjectName], v)
	}

	for namespace := range svcs {

		// Get all services and endpoints that exist in the corresponding namespace
		clusterLabelSelector := fmt.Sprintf("%s=%s", translator.GimbalLabelBackend, r.BackendName)
		currentServices, err := r.GimbalKubeClient.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: clusterLabelSelector})
		if err != nil {
			r.Metrics.GenericMetricError("ListServicesInNamespace")
			log.Errorf("error listing services in namespace %q: %v", namespace, err)
			continue
		}

		// Reconcile current state with desired state
		desiredSvcs := kubeServices(r.BackendName, svcs[namespace])
		r.reconcileSvcs(desiredSvcs, currentServices.Items)

		// Group endpoints by `Application`
		eps := map[string][]VirtualMachine{}
		for _, v := range svcs[namespace] {
			eps[v.Application] = append(eps[v.Application], v)
		}

		for app, e := range eps {
			clusterLabelSelector := fmt.Sprintf("%s=%s, %s=%s", translator.GimbalLabelBackend, r.BackendName, translator.GimbalLabelService, app)
			currentk8sEndpoints, err := r.GimbalKubeClient.CoreV1().Endpoints(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: clusterLabelSelector})
			if err != nil {
				r.Metrics.GenericMetricError("ListEndpointsInNamespace")
				log.Errorf("error listing endpoints in namespace:%q: %v", namespace, err)
				continue
			}

			// Convert the k8s list to type []Endpoint so make comparison easier
			currentEndpoints := []translator.Endpoint{}
			for _, v := range currentk8sEndpoints.Items {
				currentEndpoints = append(currentEndpoints, translator.Endpoint{Endpoints: v, UpstreamName: v.Name})
			}

			desiredEndpoints := kubeEndpoints(r.BackendName, namespace, app, e)
			r.reconcileEndpoints(desiredEndpoints, currentEndpoints)

			//// Log upstream /invalid services to prometheus
			//			//r.Metrics.DiscovererUpstreamServicesMetric(projectName, totalUpstreamServices)
			//			//r.Metrics.DiscovererInvalidServicesMetric(projectName, totalInvalidServices)
			//			//
			//			//for _, ep := range desiredEndpoints {
			//			//	totalUpstreamEndpoints := sync.SumEndpoints(&ep.Endpoints)
			//			//	r.Metrics.DiscovererUpstreamEndpointsMetric(projectName, ep.UpstreamName, totalUpstreamEndpoints)
			//			//}
			//
			//fmt.Printf("%s: %s - %s\n", vm.Summary.Config.Name, vm.Summary.Guest.IpAddress, vm.Summary.Config.Uuid)
			//
		}
	}

	// Log to Prometheus the cycle duration
	r.Metrics.CycleDurationMetric(time.Since(start))
}

func lookupTags(tags []tags.Tag) (string, string, string) {
	project := ""
	application := ""
	port := ""
	for _, t := range tags {
		switch t.Name {
		case "gimbal-project":
			project = t.Description
		case "gimbal-port":
			port = t.Description
		case "gimbal-application":
			application = t.Description
		}
	}
	return project, application, port
}

type ByName []mo.VirtualMachine

func (n ByName) Len() int           { return len(n) }
func (n ByName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n ByName) Less(i, j int) bool { return n[i].Name < n[j].Name }

func (r *Reconciler) reconcileSvcs(desiredSvcs, currentSvcs []v1.Service) {
	add, up, del := translator.DiffServices(desiredSvcs, currentSvcs)
	for _, svc := range add {
		s := svc
		r.syncqueue.Enqueue(sync.AddServiceAction(&s))
	}
	for _, svc := range up {
		s := svc
		r.syncqueue.Enqueue(sync.UpdateServiceAction(&s))
	}
	for _, svc := range del {
		s := svc
		r.syncqueue.Enqueue(sync.DeleteServiceAction(&s))
	}
}

func (r *Reconciler) reconcileEndpoints(desired, current []translator.Endpoint) {
	add, up, del := translator.DiffEndpoints(desired, current)
	for _, ep := range add {
		e := ep
		r.syncqueue.Enqueue(sync.AddEndpointsAction(&e.Endpoints, e.UpstreamName))
	}
	for _, ep := range up {
		e := ep
		r.syncqueue.Enqueue(sync.UpdateEndpointsAction(&e.Endpoints, e.UpstreamName))
	}
	for _, ep := range del {
		e := ep
		r.syncqueue.Enqueue(sync.DeleteEndpointsAction(&e.Endpoints, e.UpstreamName))
	}
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}
