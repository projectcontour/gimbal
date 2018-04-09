# Prometheus

This directory contains a sample development deployment of Prometheus and Alert Manager using temporary storage (e.g. emptyDir space).

## Quick start

```sh
# Create namespace
kubectl apply -f prometheus/deployment/namespace.yaml

# Create prometheus config
kubectl apply -f prometheus/deployment/prometheus-configmap.yaml

# Create prometheus deployment
kubectl apply -f prometheus/deployment/prometheus-deployment.yaml

# Create the prometheus node exporter
kubectl apply -f prometheus/deployment/prometheus-node-exporter.yaml
```

## Access the Prometheus Web UI

```sh
kubectl -n gimbal-monitoring port-forward $(kubectl -n gimbal-monitoring get pods -l app=prometheus -l component=server -o jsonpath='{.items[0].metadata.name}') 9090:9090
```

then go to [http://localhost:9090](http://localhost:9090) in your browser

## Access the Alert Manager Web UI

```sh
kubectl -n gimbal-monitoring port-forward $(kubectl -n gimbal-monitoring get pods -l app=prometheus -l component=alertmanager -o jsonpath='{.items[0].metadata.name}') 9093:9093
```
