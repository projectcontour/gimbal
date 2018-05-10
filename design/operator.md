# Gimbal Operator

Heptio Gimbal is an ingress load balancing platform capable of routing traffic to multiple Kubernetes and OpenStack clusters. The Gimbal Operator is designed to assist with various aspects of Gimbal from configuring and deploying Discoverers as well as managing the lifecycle of ingress traffic to clusters.

## Goals

- Deploy discoverers via CRD
- Manage cluster maintenance mode by shifting traffic from the cluster going down to other upstreams
- Verify backend clusters are configured unique names
- Expose errors from Discovery (e.g. services that are discovered but invalid in some way)
- Manage credential rotation for Discoverers

## Background

Gimbal is a comprised of many components and managing them can be difficult. The goal of the Gimbal operator is to make managing the components easier for Administrators as well as provide operational knowledge to handle events that occur within the Gimbal cluster. 

## High-level design

The primary goal of the Gimbal Operator is to manage one or more Discoverers for a Gimbal deployment. It does this by watching a Custom Resource Definition (CRD) which defines the access credentials, API Endpoint, as well as any other information needed to connect to remote clusters. When the Operator sees a new Discoverer CRD created, it will launch a new Discoverer. In the case that the CRD is changed, it will update the Discoverer and restart accordingly. In the event a CRD is deleted, then the Discoverer will be terminated and all associated ingress objects will be updated to remove reference to services related to that discovery backend.

The Gimbal operator also manages the lifecycle of the Ingress routes which reference this DiscoveryBackend. In the event the cluster needs to be taken down, the user will change the clusterStatus field to “Maintenance”. The operator will notice this change, update all Ingress routes to redirect traffic away from the backend. 

## Detailed design

The Gimbal Operator can manage the set of Discoverers for a Gimbal cluster which are responsible for discovering services/endpoints in any remote system (Current release is Kubernetes & Openstack). 

### Discovery Backend

To deploy a discovery backend, an Administrator collects the required information for the discovery backend and creates a secret in the `gimbal-discovery` namespace. Next a CRD is created which informs the Operator that a new Discovery Backend needs to be provisioned. The cluster-name parameter of the Discoverer is derived from the name field of the DiscoveryBackend CRD. Since the CRDs are deployed within the same namespace, the cluster-name is then guaranteed to be unique. 

### Credential Rotation

Each discoverer requires various credentials which are stored in Kubernetes secrets. Upon startup, the operator reads the secret defined in the CRD and initializes the appropriate Discoverer. If credentials need to be rotated the Administrator would update the existing secret, the operator would see that change and apply the new configuration by restarting the Discoverer. An alternative would be to create a new secret and update the CRD to reference the new secret. (Future: Schedule the credential rotation for the future so it happens off-hours?). 

### ClusterStatus

The ClusterStatus field of the CRD tells the operator if the Discovery Backend is available to take requests or not. This is an Administrator managed field and is used in the case where the backend needs to be taken out of service for maintenance or some other reason. When the field changes from `Enabled` to `Maintenance`, the operator will look for all the Ingress resources which match the discovery backend and scale their weight to zero as well as store the existing weight configuration in the Ingress object. This allows the Ingress resources to retain references to services matching the discovery backend. Once maintenance is complete, the Administrator changes the field back to “Enabled” and restores the previous weighting before the cluster was taken offline. If weighting was not defined previously, then even distribution is assumed.

### Errors

Various errors can occur during startup and operation of a Discoverer. Those errors are logged via Prometheus metrics as well as the related Discovery CRD has an Errors section which can be populated. This gives Administrators an easy way to find errors in a central way without needing to tail logs on a pod or setup a centralized logging system. 

### CRD

Custom resource definition overview:

- apiVersion: gimbal.heptio.com/v1
- kind: DiscoveryBackend
- metadata:
  - namespace: Namespace to deploy the Discoverer into 
  - name: Name of the backend cluster (Must be lowercase)
- cluster:
  - type: openstack or kubernetes
  - secret: name of secret containing extra information to connect to backend cluster
  - refreshInterval (seconds): Interval in seconds to cause a new cycle 
  - debug (bool): Enable debug logging for troubleshooting
  - clusterStatus (Enabled, Maintenance): Defines if the cluster is accepting connections for down for maintenance
- status:
  - []errors: Errors encountered during discovery
    - statusTime: Time event was recorded
    - type: Reason for error
    - message: More detailed error message

#### Example

```yaml
apiVersion: gimbal.heptio.com/v1
kind: Discoverer
metadata:
  namespace: gimbal-discovery
  name: openstack001
cluster:
  type: openstack
  secret: remote-discover-openstack 
  refreshInterval: 30
  clusterStatus: enabled
---
apiVersion: contour.heptio.com/v1
kind: Discovery
metadata:
  namespace: gimbal-discovery
  name: kubernetes001
cluster:
  type: kubernetes
  credentials: kubernetes001-kubecfg
  clusterStatus: enabled
```