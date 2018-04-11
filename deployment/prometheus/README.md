# Prometheus

This directory contains a sample development deployment of Prometheus and Alert Manager using temporary storage (e.g. emptyDir space).

## Quick start

```sh
# Navigate to directory starting at root
$ cd deployment/prometheus

# Create namespace (If required)
$ kubectl apply -f deployment/01-namespace.yaml

# Create prometheus config
$ kubectl apply -f deployment/02-prometheus-configmap.yaml

# Create alert manager config
$ kubectl apply -f deployment/02-prometheus-alertmanager-configmap.yaml

# Create alert rules config
$ kubectl apply -f deployment/02-prometheus-alertrules-configmap.yaml

# Create prometheus deployment
$ kubectl apply -f deployment/03-prometheus-deployment.yaml

# Create alertmanager deployment
$ kubectl apply -f deployment/03-prometheus-alertmanager-deployment.yaml

# Create the prometheus node exporter
$ kubectl apply -f deployment/03-prometheus-node-exporter.yaml

# Deploy kube-state-metrics to gather cluster information
$ git clone https://github.com/kubernetes/kube-state-metrics.git
$ kubectl apply -f kubernetes
```

## Access the Prometheus Web UI

```sh
$ kubectl -n gimbal-monitoring port-forward $(kubectl -n gimbal-monitoring get pods -l app=prometheus -l component=server -o jsonpath='{.items[0].metadata.name}') 9090:9090
```

then go to [http://localhost:9090](http://localhost:9090) in your browser

## Access the Alert Manager Web UI

```sh
$ kubectl -n gimbal-monitoring port-forward $(kubectl -n gimbal-monitoring get pods -l app=prometheus -l component=alertmanager -o jsonpath='{.items[0].metadata.name}') 9093:9093
```

then go to [http://localhost:9093](http://localhost:9093) in your browser
