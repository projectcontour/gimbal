# Kubernetes Discoverer

## Overview

The Kubernetes discoverer provides service discovery for a Kubernetes cluster. It does this by monitoring available Services and Endpoints for a single Kubernetes cluster and synchronizing them into the host Gimbal cluster.

The Discoverer will leverage the watch feature of the Kubernetes API to receive changes dynamically, rather than having to poll the API. All available services & endpoints will be synchronized to the same namespace matching the source system.

The discoverer will only be responsible for monitoring a single cluster at a time. If multiple clusters are required to be watched, then multiple discoverers will need to be deployed.

## Technical Details

The following sections outline the technical implementations of the discoverer.

See the [design documentation](../discovery/design/kubernetes.md) for additional details.

### Arguments

Arguments are available to customize the discoverer, most have defaults but others are required to be configured by the cluster administrators:

| flag  | default  | description  |
|---|---|---|
| version  |  false | Show version, build information and quit  
| num-threads  | 2  |  Specify number of threads to use when processing queue items
| gimbal-kubecfg-file  | ""  | Location of kubecfg file for access to Kubernetes cluster hosting Gimbal
| discover-kubecfg-file | ""  | Location of kubecfg file for access to remote Kubernetes cluster to watch for services / endpoints 
| backend-name  | ""  |   Name of cluster scraping for services & endpoints (Cannot start or end with a hyphen and must be lowercase alpha-numeric)
| debug | false | Enable debug logging 
| prometheus-listen-address | 8080 | The address to listen on for Prometheus HTTP requests
| gimbal-client-qps | 5 | The maximum queries per second (QPS) that can be performed on the Gimbal Kubernetes API server
| gimbal-client-burst | 10 | The maximum number of queries that can be performed on the Gimbal Kubernetes API server during a burst

### Credentials

A valid [Kubernetes config file](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) is required to access the remote cluster. This config file is stored as a Kubernetes secret in the Gimbal cluster.

The following example creates a secret from a file locally and places it in the namespace `gimbal-discovery`. **_NOTE: Path to `config file` as well as `backend-name` will need to be customized._**

```bash
# Sample secret creation
$ kubectl create secret generic remote-discover-kubecfg --from-file=./config --from-literal=backend-name=nodek8s -n gimbal-discovery
```

#### Sample Kubernetes config file

Sample configuration file for a Kubernetes cluster:

```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /Users/stevesloka/.minikube/ca.crt
    server: https://192.168.64.13:8443
  name: node01
contexts:
- context:
    cluster: node01
    user: node01
  name: node01
current-context: node01
kind: Config
preferences: {}
users:
- name: node01
  user:
    client-certificate-data: <base64>
    client-key-data: <base64>
```

### Updating Credentials

Credentials to the backend Kubernetes cluster can be updated at any time if necessary. To do so, we recommend taking advantage of the Kubernetes deployment's update features:

1. Create a new secret with the new credentials.
2. Update the deployment to reference the new secret.
3. Wait until the discoverer pod is rolled over.
4. Verify the discoverer is up and running.
5. Delete the old secret, or rollback the deployment if the discoverer failed to start.

### Configuring the Gimbal Kubernetes client rate limiting

The discoverer has two configuration parameters that control the request rate limiter of the Kubernetes client used to sync services and endpoints to the Gimbal cluster:

* Queries per second (QPS): Number of requests per second that can be sent to the Gimbal API server. Set using the `--gimbal-client-qps` command-line flag.
* Burst size: Number of requests that can be sent during a burst period. A burst is a period of time in which the number of requests can exceed the configured QPS, while still maintaining a smoothed QPS rate over time. Set using the `--gimbal-client-burst` command-line flag.

These configuration parameters are dependent on your requirements and the hardware running the Gimbal cluster. If services and endpoints in your environment undergo a high rate of change, increase the QPS and burst parameters, but make sure that the Gimbal API server and etcd cluster can handle the increased load.

### Data flow

Data flows from the remote cluster into the Gimbal cluster. The steps on how they replicate are as follows:

1. Connection is made to remote cluster and all services and corresponding endpoints are retrieved from the cluster
2. Those objects are then synchronized to the Gimbal cluster in the same namespace as the remote cluster. For example, if a service named `testsvc01` exists in the namespace `team1` then the same service will be written to the Gimbal cluster in the `team1` namespace. Labels will also be added during the synchronization (See the [labels](#labels) section for more details).
3. Once the initial list of objects is synchronized, any further updates will happen automatically when a service or endpoint is `created`, `updated`, or `deleted`.

#### Ignored Objects

An exception to the flow outlined previously are objects that are ignored when synchronizing. The following rules determine if an object is ignored during sync:

- Any service or endpoint in the `kube-system` namespace
- Any service or endpoint named `kubernetes` in the `default` namespace

### Labels

All synchronized services & endpoints will contain the same properties as the source system (e.g. annotations, labels, etc), but additional labels are added to assist in understanding where the object was sourced from.

Labels added to service and endpoints:
```
gimbal.heptio.com/service=<serviceName>
gimbal.heptio.com/backend=<nodeName>
```