<!-- TOC -->

- [Deployment](#deployment)
    - [Setup / Requirements](#setup--requirements)
    - [Contour](#contour)
    - [Discoverers](#discoverers)
        - [Kubernetes](#kubernetes)
        - [Openstack](#openstack)
    - [Prometheus](#prometheus)
        - [Quick start](#quick-start)
        - [Access the Prometheus Web UI](#access-the-prometheus-web-ui)
        - [Access the Alert Manager Web UI](#access-the-alert-manager-web-ui)
    - [Grafana](#grafana)
        - [Quick Start](#quick-start)
        - [Accessing Grafana UI](#accessing-grafana-ui)
        - [Configure Grafana](#configure-grafana)
            - [Configure Datasource](#configure-datasource)
            - [Add Dashboard](#add-dashboard)
                - [Sample Kubernetes Dashboard](#sample-kubernetes-dashboard)
                - [Sample Envoy Dashboard](#sample-envoy-dashboard)
    - [Validation](#validation)
        - [Discovery Cluster](#discovery-cluster)
        - [Gimbal Cluster](#gimbal-cluster)

<!-- /TOC -->
# Deployment

Following are instructions to get all the components up and running. 

## Setup / Requirements

A copy of this repo locally which is easily accomplished by cloning or downloading a copy: 

```sh
$ git clone git@github.com:heptio/gimbal.git
```

Additionally, this guide will assume a single Kubernetes cluster to deploy Gimbal, as well as Kubernetes or Openstack clusters with `flat` networking, meaning, pods get route-able IP address on the network.

## Contour

Contour is the system which handles the Ingress traffic. It utilizes Envoy which is an L7 proxy and communication bus designed for large modern service oriented architectures. 

Envoy is the data component of Contour and handles the network traffic, Contour drives the configuration of Envoy based upon the Kubernetes cluster configuration.

```sh
# Navigate to deployment directory
$ cd deployment

# Deploy Contour
$ kubectl create -f contour/
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
$ kubectl create secret generic remote-discover-kubecfg --from-file=./config --from-literal=cluster-name=node02 -n gimbal-discoverer

# Deploy Discoverer
$ kubectl apply -f gimbal-discoverer/02-kubernetes-discoverer.yaml
```

Technical details on how the Kubernetes Discoverer works can be found in the [docs section](../docs/discoverer/kubernetes/README.md).

### Openstack

The Openstack Discoverer is responsible for looking at all LBaSS and members in an Openstack cluster and synchronizing them to the host cluster. 
 
[Credentials](../docs/discoverer/openstack/README.md#credentials) to the remote cluster are required to be created as a secret. 

```
# Openstack secret
$ kubectl create secret generic remote-discover-openstack --from-file=certificate-authority-data=./ca.pem --from-literal=cluster-name=openstack --from-literal=username=admin --from-literal=password=abc123 --from-literal=auth-url=https://api.openstack:5000/ --from-literal=tenant-name=heptio

# Deploy Discoverer
$ kubectl apply -f gimbal-discoverer/02-openstack-discoverer.yaml
```

Technical details on how the Openstack Discoverer works can be found in the [docs section](../docs/discoverer/openstack/README.md).

## Prometheus

This directory contains a sample development deployment of Prometheus and Alert Manager using temporary storage (e.g. emptyDir space).

### Quick start

```sh
# Deploy 
$ kubectl apply -f prometheus

# Deploy kube-state-metrics to gather cluster information
$ git clone https://github.com/kubernetes/kube-state-metrics.git
$ cd kube-state-metrics
$ kubectl apply -f kubernetes/
```

### Access the Prometheus Web UI

```sh
$ kubectl -n gimbal-monitoring port-forward $(kubectl -n gimbal-monitoring get pods -l app=prometheus -l component=server -o jsonpath='{.items[0].metadata.name}') 9090:9090
```

then go to [http://localhost:9090](http://localhost:9090) in your browser

### Access the Alert Manager Web UI

```sh
$ kubectl -n gimbal-monitoring port-forward $(kubectl -n gimbal-monitoring get pods -l app=prometheus -l component=alertmanager -o jsonpath='{.items[0].metadata.name}') 9093:9093
```

then go to [http://localhost:9093](http://localhost:9093) in your browser

## Grafana

Sample development deployment of Grafana using temporary storage.

### Quick Start

```sh
# Deploy
$ kubectl apply -f grafana/

# Create secret with grafana credentials
$ kubectl create secret generic grafana -n gimbal-monitoring \
    --from-literal=grafana-admin-password=admin \
    --from-literal=grafana-admin-user=admin 
```

### Accessing Grafana UI

```sh
$ kubectl port-forward $(kubectl get pods -l app=grafana -n gimbal-monitoring -o jsonpath='{.items[0].metadata.name}') 3000 -n gimbal-monitoring
```

Access Grafana at http://localhost:3000 in your browser. Use `admin` as the username.

### Configure Grafana

Grafana requires some configuration after it's deployed, use the following steps to configure a datasource and import a dashboard to validate the connection. 

#### Configure Datasource

1. From main Grafana page, click on `Add Datasource`
2. For `Name` enter `prometheus`
3. Choose `Prometheus` under `Type` selector
4. In the next section, enter `http://prometheus:9090` for the `URL`
5. Click `Save & Test`
6. Look for the message box in green stating `Data source is working`

#### Add Dashboard

##### Sample Kubernetes Dashboard

Add sample dashboard to validate data source is collecting data:

1. From main page, click on `plus` icon and choose `Import dashboard`
2. Enter `1621` in the first box
3. Under the `prometheus` section choose the data source created in previous section
4. Click `Import`

##### Sample Envoy Dashboard

The `dashboards/` directory contains sample dashboards for monitoring Envoy metrics with Grafana. The dashboards are
JSON-encoded, and can be imported into Grafana as follows:

1. From main page, click on `plus` icon and choose `Import dashboard`
2. Click on `Upload .json File`
3. Navigate to the `dashboards/` directory and select the `envoy-metrics.json` file
4. Select Prometheus as the datasource
5. Click `Import`

## Validation

Once the components are deployed, the deployment can be verified with the following steps:

### Discovery Cluster

This example utilizes a Kubernetes cluster as the discovered cluster which was configured [previously](#kubernetes). We will deploy a few sample applications into the `default` namespace:

```sh
# Deploy sample apps
$ kubectl apply -f example-workload/deployment.yaml
```

### Gimbal Cluster

These commands should be run on the Gimbal cluster to verify it's components:

```sh
# Verify Discoverer Components
$ kubectl get po -n gimbal-discoverer
NAME                                         READY     STATUS    RESTARTS   AGE
k8s-kubernetes-discoverer-55899dcb66-lgvnk   1/1       Running   0          5m

# Verify Contour
$ kubectl get po -n heptio-contour
NAME            READY     STATUS    RESTARTS   AGE
contour-lq6mm   2/2       Running   0          5h

# Verify discovered services
$ kubectl get svc -l gimbal.heptio.com/cluster=node02 
NAME                TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
nginx-node02        ClusterIP   None         <none>        80/TCP    17m
kuard-node02        ClusterIP   None         <none>        80/TCP    17m

# Deploy an Ingress route
$ kubectl apply -f example-workload/ingress.yaml

# Port forward to the Contour pod
$ kubectl port-forward $(kubectl get pods -n heptio-contour -l app=contour -o jsonpath='{.items[0].metadata.name}') 9000:80 -n heptio-contour

# Make a request to Gimbal cluster which will proxy traffic to the secondary cluster
$ curl -i -H "Host: kuard.local" localhost:9000
$ curl -i -H "Host: nginx.local" localhost:9000
```
