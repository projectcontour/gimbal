# Grafana

Sample development deployment of Grafana using temporary storage.

## Quick Start

```sh

# Set admin password
GRAFANA_PASSWORD=

# Create secret with grafana credentials
kubectl create secret generic -n contour-monitoring \
    --from-literal=grafana-admin-password=${GRAFANA_PASSWORD} \
    --from-literal=grafana-admin-user=admin

# Apply resources
kubectl apply -f deployment/
```

## Accessing Grafana UI

```sh
kubectl port-forward $(kubectl get pods -l app=grafana -n contour-monitoring -o jsonpath='{.items[0].metadata.name}') 3000
```

Access Grafana at http://localhost:3000 in your browser. Use `admin` as the username.
