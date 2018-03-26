# Prometheus

This directory contains a sample development deployment of Prometheus using temporary storage (e.g. emptyDir space)

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
kubectl -n contour-monitoring port-forward $(kubectl get pods -l app=prometheus -n contour-monitoring -o jsonpath='{.items[0].metadata.name}') 9090:9090
```

then go to http://localhost:9090 in your browser
