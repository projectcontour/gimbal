# Openstack Discoverer

## Overview

The Openstack discoverer provides service discovery for an Openstack cluster. It does this by monitoring all Load Balancer as a Service (LBaaS) configured as well as the corresponding Members. They are synchronized to the Team namespace as Services and Endpoints, with the Namespace being configured as the TenantName in Openstack.

The Discoverer will poll the Openstack API on a customizable interval and update the Contour cluster accordingly.

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
| cluster-name  | ""  |   Name of cluster scraping for services & endpoints 
| debug | false | Enable debug logging 
| reconciliation-period | 30s | The interval of time between reconciliation loop runs 
| http-client-timeout | 5s | The HTTP client request timeout
| openstack-certificate-authority | "" | Path to cert file of the OpenStack API certificate authority

### Credentials

The discoverer requires a username/password, auth URL, as well as the TenantName of the Openstack cluster to be discovered:

Credentials required:
- Username: User with access to api
- Password: Password for User
- AuthURL: Openstack API Url
- TenantName: Tenant used to discover services

_NOTE: These are exposed to the deployment via environment variables._

### Data flow

Data flows from the remote cluster into the Contour cluster. The steps on how they replicate are as follows:

1. Connection is made to remote cluster and all LBaaS's and corresponding Members are retrieved from the cluster
2. Those objects are then translated into Kubernetes Services and Endpoints, then synchronized to the Contour cluster in the same namespace as the remote cluster. Labels will also be added during the synchronization (See the [labels](#labels) section for more details).
3. Once the initial list of objects is synchronized, any further updates will happen based upon the configured `reconciliation-period` which will start a new reconciliation loop.

### Labels

All synchronized services & endpoints will have additional labels added to assist in understanding where the object were sourced from. 

Labels added to service and endpoints:
```
contour.heptio.com/service=<serviceName>
contour.heptio.com/cluster=<nodeName>
contour.heptio.com/load-balancer-id=<LoadBalancer.ID>
contour.heptio.com/load-balancer-name=<LoadBalancer..Name>
```