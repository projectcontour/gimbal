# Deployment

Following are instructions to get all the components up and running.

## Contour

Contour is the system which handles the Ingress traffic. It utilizes Envoy which is an L7 proxy and communication bus designed for large modern service oriented architectures. 

Envoy is the data component of Contour and handles the network traffic, Contour drives the configuration of Envoy based upon the Kubernetes cluster configuration. 

```
# Create namespace / service account
$ kubectl create -f contour/01-common.yaml

# Create RBAC policies for Service Account
$ kubectl create -f contour/02-rbac.yaml

# Deploy Contour
$ kubectl apply -f contour/02-contour.yaml

# Deploy Contour Service
$ kubectl apply -f contour/02-service.yaml
```

_NOTE: The current configuration exposes the Envoy Admin UI so that Prometheus can scrape for metrics!_

## Discoverers

Service discovery is enabled via the Discoverers which have both Kubernetes and Openstack implementations.

```
# Create contour-discoverer namespace
kubectl create -f contour-discoverer/01-common.yaml
```

### Kubernetes

The Kubernetes Discoverer is responsible for looking at all services and endpoints in a Kubernetes cluster and synchronizing them to the host cluster. 

[Credentials](../docs/discoverer/kubernetes/README.md#credentials) to the remote cluster are required to be created as a secret. 

```
# Kubernetes secret
$ kubectl create secret generic remote-discover-kubecfg --from-file=./config -n contour-discoverer

# Deploy Discoverer
$ kubectl apply -f contour-discoverer/02-kubernetes-discoverer.yaml
```

Technical details on how the Kubernetes Discoverer works can be found in the [docs section](../docs/discoverer/kubernetes/README.md).

### Openstack

The Openstack Discoverer is responsible for looking at all LBaSS and members in an Openstack cluster and synchronizing them to the host cluster. 
 
[Credentials](../docs/discoverer/kubernetes/README.md#credentials) to the remote cluster are required to be created as a secret. 

```
# Deploy Discoverer
$ kubectl apply -f contour-discoverer/02-openstack-discoverer.yaml
```

Technical details on how the Openstack Discoverer works can be found in the [docs section](../docs/discoverer/openstack/README.md).

## Validation

Once the components are deployed, the deployment can be verified with the following steps:

```
# Verify Discoverer Components
$ kubectl get po -n contour-discoverer

# Verify Contour
$ kubectl get po -n heptio-contour

# Deploy a Route CRD
$ kubectl apply -f example-workload/route.yaml

# Port forward to the contour pod
$ kubectl port-forward $(kubectl get pods -n heptio-contour -l app=contour -o jsonpath='{.items[0].metadata.name}') 9000:80 -n heptio-contour

# Make a request
$ curl -i -H "Host: kuard.local" localhost:9000
$ curl -i -H "Host: nginx.local" localhost:9000
```
