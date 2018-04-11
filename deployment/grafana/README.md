# Grafana

Sample development deployment of Grafana using temporary storage.

## Quick Start

```sh

# Navigate to grafana directory from root
$ cd deployment/grafana

# Apply Namespace (if required)
$ kubectl apply -f deployment/01-namepace.yaml

# Set admin password
$ GRAFANA_PASSWORD=admin

# Create secret with grafana credentials
$ kubectl create secret generic grafana -n gimbal-monitoring \
    --from-literal=grafana-admin-password=${GRAFANA_PASSWORD} \
    --from-literal=grafana-admin-user=admin 

# Apply Configmap
$ kubectl apply -f deployment/02-grafana-configmap.yaml

# Apply Deployment
$ kubectl apply -f deployment/03-grafana-deployment.yaml

# Apply Configmap
$ kubectl apply -f deployment/03-grafana-service.yaml
```

## Accessing Grafana UI

```sh
$ kubectl port-forward $(kubectl get pods -l app=grafana -n gimbal-monitoring -o jsonpath='{.items[0].metadata.name}') 3000 -n gimbal-monitoring
```

Access Grafana at http://localhost:3000 in your browser. Use `admin` as the username.

## Configure Grafana

Grafana requires some configuration after it's deployed, use the following steps to configure a datasource and import a dashboard to validate the connection. 

### Configure Datasource

1. From main Grafana page, click on `Add Datasource`
2. For `Name` enter `prometheus`
3. Choose `Prometheus` under `Type` selector
4. In the next section, enter `http://prometheus:9090` for the `URL`
5. Click `Save & Test`
6. Look for the message box in green stating `Data source is working`

### Add Dashboard

#### Sample Kubernetes Dashboard

Add sample dashboard to validate data source is collecting data:

1. From main page, click on `plus` icon and choose `Import dashboard`
2. Enter `1621` in the first box
3. Under the `prometheus` section choose the data source created in previous section
4. Click `Import`

#### Sample Envoy Dashboard

The `dashboards/` directory contains sample dashboards for monitoring Envoy metrics with Grafana. The dashboards are
JSON-encoded, and can be imported into Grafana as follows:

1. From main page, click on `plus` icon and choose `Import dashboard`
2. Click on `Upload .json File`
3. Navigate to the `dashboards/` directory and select the `envoy-metrics.json` file
4. Select Prometheus as the datasource
5. Click `Import`
