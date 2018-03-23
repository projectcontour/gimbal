# Kubernetes Discoverer

## Overview 

The Kubernetes discoverer provides service discovery for a Kubernetes cluster. It does this by monitoring available Services and Endpoints for a single Kubernetes cluster and synchronizing them into the host Contour cluster. 

The Discoverer will leverage the watch feature of the Kubernetes API to receive changes dynamically, rather than having to poll the API. All available services & endpoints will be synchronized to the the same namespace matching the source system.

The discoverer will only be responsible for monitoring a single cluster at a time. If multiple clusters are required to be watched, then multiple discoverer controllers will need to be deployed. 

## Technical Details

The following sections outline the technical implementations of the discoverer.

### Arguments

Arguments are available to customize the discoverer, most have defaults but others are required to be configured by the cluster administrators:

| flag  | default  | description  |
|---|---|---|
| version  |  false | Show version, build information and quit  
| num-threads  | 2  |  Specify number of threads to use when processing queue items
| contour-kubecfg-file  | ""  | Location of kubecfg file for access to Kubernetes cluster hosting Contour
| discover-kubecfg-file | ""  | Location of kubecfg file for access to remote Kubernetes cluster to watch for services / endpoints 
| cluster-name  | ""  |   Name of cluster scraping for services & endpoints 
| debug | false | Enable debug logging 

### Credentials

A valid [Kubernetes config file](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) is required to access the remote cluster. This config file is stored as a Kubernetes secret in the Contour cluster. 

The following example creates a secret from a file locally and places it in the namespace `contour-discoverer`: 

```
# -- Sample secret creation
$ kubectl create secret generic remote-discover-kubecfg --from-file=./config -n contour-discoverer
```

#### Sample Kubernetes config file

Sample configuration file for a Kubernetes cluster:

```
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

### Data flow

Data flows from the remote cluster into the Contour cluster. The steps on how they replicate are as follows:

1. Connection is made to remote cluster and all services and corresponding endpoints are retrieved from the cluster
2. Those objects are then synchronized to the Contour cluster in the same namespace as the remote cluster. For example, if a service named `testsvc01` exists in the namespace `team1` then the same service will be written to the Contour cluster in the `team1` namespace. Labels will also be added during the synchronization (See the [labels](#labels) section for more details).
3. Once the initial list of objects is synchronized, any further updates will happen automatically when a service or endpoint is `created`, `updated`, or `deleted`.

### Labels

All synchronized services & endpoints will contain the same properties as the source system (e.g. annotations, labels, etc), but additional labels are added to assist in understanding where the object was sourced from. 

Labels added to service and endpoints:
```
contour.heptio.com/service=<serviceName>
contour.heptio.com/cluster=<nodeName>
```