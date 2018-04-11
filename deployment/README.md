# Deployment

Following are instructions to get all the components up and running. 

## Setup

Clone this repo: 

```sh
$ git clone git@github.com:heptio/gimbal.git
```

## Contour

Contour is the system which handles the Ingress traffic. It utilizes Envoy which is an L7 proxy and communication bus designed for large modern service oriented architectures. 

Envoy is the data component of Contour and handles the network traffic, Contour drives the configuration of Envoy based upon the Kubernetes cluster configuration. 

```sh
# Navigate to deployment directory
$ cd deployment

# Create namespace / service account
$ kubectl create -f contour/01-common.yaml

# Create RBAC policies for Service Account
$ kubectl create -f contour/02-rbac.yaml

# Deploy Contour
$ kubectl apply -f contour/02-deployment.yaml

# Deploy Contour Service
$ kubectl apply -f contour/02-service.yaml
```

_NOTE: The current configuration exposes the Envoy Admin UI so that Prometheus can scrape for metrics!_

## Discoverers

Service discovery is enabled via the Discoverers which have both Kubernetes and Openstack implementations.

```
# Create gimbal-discoverer namespace
kubectl create -f gimbal-discoverer/01-common.yaml
```

### Kubernetes

The Kubernetes Discoverer is responsible for looking at all services and endpoints in a Kubernetes cluster and synchronizing them to the host cluster. 

[Credentials](../docs/discoverer/kubernetes/README.md#credentials) to the remote cluster are required to be created as a secret. 

```
# Kubernetes secret
$ kubectl create secret generic remote-discover-kubecfg --from-file=./config --from-literal=cluster-name=nodek8s -n gimbal-discoverer

# Deploy Discoverer
$ kubectl apply -f gimbal-discoverer/02-kubernetes-discoverer.yaml
```

Technical details on how the Kubernetes Discoverer works can be found in the [docs section](../docs/discoverer/kubernetes/README.md).

### Openstack

The Openstack Discoverer is responsible for looking at all LBaSS and members in an Openstack cluster and synchronizing them to the host cluster. 
 
[Credentials](../docs/discoverer/openstack/README.md#credentials) to the remote cluster are required to be created as a secret. 

```
# Openstack secret: Fill out the following "data" values and verify they are base64 encoded:

apiVersion: v1
kind: Secret
metadata:
  name: remote-discover-openstack
  namespace: gimbal-discoverer
type: Opaque
data:
  cluster-name: clustername_base64
  username: username_base64
  password: password_base64
  auth-url: authurl_base64
  tenant-name: tenantname_base64
  certificate-authority-data: certdata_base64

# Deploy Discoverer
$ kubectl apply -f gimbal-discoverer/02-openstack-discoverer.yaml
```

Technical details on how the Openstack Discoverer works can be found in the [docs section](../docs/discoverer/openstack/README.md).

## Metrics

Prometheus & Grafana are used to manage performance metrics and statistics for all systems. Deployment / setup configurations can be found here:

- [Prometheus](prometheus/README.md)
- [Grafana](grafana/README.md)

## Validation

Once the components are deployed, the deployment can be verified with the following steps:

```
# Verify Discoverer Components
$ kubectl get po -n gimbal-discoverer

# Verify Contour
$ kubectl get po -n heptio-contour

# Deploy an Ingress route
$ kubectl apply -f example-workload/ingress.yaml

# Port forward to the Contour pod
$ kubectl port-forward $(kubectl get pods -n heptio-contour -l app=contour -o jsonpath='{.items[0].metadata.name}') 9000:80 -n heptio-contour

# Make a request
$ curl -i -H "Host: kuard.local" localhost:9000
$ curl -i -H "Host: nginx.local" localhost:9000
```
