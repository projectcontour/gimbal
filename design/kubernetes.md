# Kubernetes Discoverer

## Executive Summary

In order for an Ingress controller to route traffic to endpoints off-cluster, we need to replicate Service and Endpoint information from the target cluster into the Gimbal cluster. The Kubernetes Discoverer establishes watches for Service and Endpoint objects in a target Kubernetes cluster and writes a copy of the object into the Gimbal cluster.

## Goals
- Maintain a valid list of Services & Endpoints from a single Kubernetes cluster replicated to corresponding namespaces in the Gimbal cluster
- Utilize `Watch` for changes to Services & Endpoint resources to avoid polling the cluster

## Non-goals

- Monitor more than one target cluster. Each target cluster will have its own dedicated Kubernetes Discoverer deployment.

## Definitions

- *Service*: A set of similar Endpoints that belong to a single Kubernetes Service.
- *Endpoint*: An IP:Port pair that correspond to a pod that can receive HTTP traffic.

### Service Example

```
Name:              nginx-node02
Namespace:         team1
Labels:            gimbal.heptio.com/service=nginx
                   gimbal.heptio.com/backend=node02
                   run=nginx
Annotations:       <none>
Selector:          <none>
Type:              ClusterIP
IP:                None
Port:              <unset>  80/TCP
TargetPort:        80/TCP
Endpoints:         172.17.0.10:80,172.17.0.11:80,172.17.0.12:80 + 2 more...
Session Affinity:  None
Events:            <none>
```

### Endpoint Example

```
Name:         nginx-node02
Namespace:    team1
Labels:       gimbal.heptio.com/service=nginx
              gimbal.heptio.com/backend=node02
              run=nginx
Annotations:  <none>
Subsets:
  Addresses:          172.17.0.10,172.17.0.11,172.17.0.12,172.17.0.4,172.17.0.9
  NotReadyAddresses:  <none>
  Ports:
    Name     Port  Protocol
    ----     ----  --------
    <unset>  80    TCP

Events:  <none>
```

## Background

The Kubernetes Discoverer responsible for monitoring available Services and Endpoints for a single Kubernetes cluster. The credentials to access the Kubernetes API of the target cluster will be provided by the Administrators via a Kubernetes Secret.

The discoverer will start a `watch` on `services` & `endpoints` on a remote Kubernetes cluster. From there it will sync those objects to a `working namespace` inside the host Kubernetes cluster so that those can be picked up by Gimbal. 

Based on the Service and Pod definitions, Endpoints’ "Target Port" may be different than tcp:80. The Kubernetes Discoverer will leverage the `watch` feature of the Kubernetes API to receive changes dynamically, rather than having to poll the API.

The Kubernetes Discoverer will write the available Services and Endpoint information to the corresponding Team namespace as standard Kubernetes services & endpoints.

The discoverer will only be responsible for monitoring a single cluster at a time. If multiple clusters are required to be watched, then multiple discoverer components will need to be deployed. Initially, discoverer's will be deployed manually via Deployments, but further iterations will introduce a Discovery Operator which will take over this responsibility. 

## Detailed Design

Watches are setup to monitor for changes to Services or Endpoints in a Kubernetes cluster. These updates (e.g. ADD, MODIFY, DELETE) are places onto a queue. Items are processed off of that rate-limited queue so that we can maintain sane performance. The queue is an in-memory queue with no durability aimed at reducing the load on the Gimbal Kubernetes API. 

If additional processing is required, then an argument can be changed to increase the number of threads available to consume items from the queue. All objects synchronized to the corresponding Team namespace. When an object is added or updated, the entire object is copied with all labels, annotations, etc. Additional labels are added so that Izzy can have more detailed information about the object. 

Those labels are defined as:

- gimbal.heptio.com/backend: BackendName (Defined via argument)
- gimbal.heptio.com/service: [ServiceName]

The name of the synchronized object will be a hash of the `BackendName-ServiceName`. The name is hashed because of length restrictions when creating a service object.

In the event that the total length of the hash is larger than 63 characters (maximum allowed length), then the components are hashed to keep within the limits. All attempts are made to keep the name as descriptive to the source as possible. 

The discoverer should synchronize the cluster on first startup so that any changes missed while being offline are properly updated. This is handled automatically since a new watch on a resource sends the current list of items upon initialization. The add logic then checks to see if the object is already existing and in that case passes it off to the update method. 

The discoverer component will have arguments which will allow users to customize or override default values:

- *num-threads*: Specify number of threads to use when processing queue items.
- *gimbal-kubecfg-file*: Location of kubecfg file for access to kubernetes cluster hosting Izzy
- *discover-kubecfg-file*: Location of kubecfg file for access to remote kubernetes cluster to watch for services / endpoints
- *backend-name*: Name of cluster scraping for services & endpoints

By default, the `kube-system` namespace is ignored when looking for services / endpoints. Future iterations of discoverer should allow for namespace whitelisting and blacklisting to allow for further customization. 

## Security/Performance Concerns

### Endpoints

Endpoints are required to be watched across multiple clusters. Since those can change frequently this can be a potential bottle neck to stay up to date as well as potentially overwhelm the host cluster since those endpoints are replicated.

# Open Issues

- Impact of watching endpoints on a large cluster

# Future

Further enhancements to the Discoverer component are to add support for monitoring Openstack clusters. 
