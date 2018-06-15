# Openstack Discoverer

## Overview

The Openstack discoverer provides service discovery for an Openstack cluster. It does this by monitoring all Load Balancer as a Service (LBaaS) configured as well as the corresponding Members. They are synchronized to the Team namespace as Services and Endpoints, with the Namespace being configured as the TenantName in Openstack.

The Discoverer will poll the Openstack API on a customizable interval and update the Gimbal cluster accordingly.

The discoverer will only be responsible for monitoring a single cluster at a time. If multiple clusters are required to be watched, then multiple discoverers will need to be deployed.

## Naming Requirements

Named OpenStack Load Balancers must have names that are compatible with the Kubernetes service naming rules.

The following requirements apply to OpenStack Load Balancers that have a non-empty name:

* The allowed characters are `A-Z`, `a-z`, `0-9` and `-`.
* The name must end with an alpha-numeric character

The OpenStack discoverer will skip any Load Balancers that do not adhere to
these rules, and log a warning that includes details about the Load Balancer
that was skipped. Additionally, the OpenStack discoverer will increment the
`gimbal_discoverer_error_total[errortype=InvalidLoadBalancerName]` prometheus
metric.

See the [naming conventions documentation](./discovery-naming-conventions.md)
for more details.

## Technical Details

The following sections outline the technical implementations of the discoverer.

See the [design document](../discovery/design/openstack.md) for additional
details.

### Arguments

Arguments are available to customize the discoverer, most have defaults but others are required to be configured by the cluster administrators:

| flag  | default  | description  |
|---|---|---|
| version  |  false | Show version, build information and quit  
| num-threads  | 2  |  Specify number of threads to use when processing queue items
| gimbal-kubecfg-file  | ""  | Location of kubecfg file for access to Kubernetes cluster hosting Gimbal
| backend-name  | ""  |   Name of cluster scraping for services & endpoints (Cannot start or end with a hyphen and must be lowercase alpha-numeric)
| debug | false | Enable debug logging 
| reconciliation-period | 30s | The interval of time between reconciliation loop runs 
| http-client-timeout | 5s | The HTTP client request timeout
| openstack-certificate-authority | "" | Path to cert file of the OpenStack API certificate authority
| prometheus-listen-address | 8080 | The address to listen on for Prometheus HTTP requests
| gimbal-client-qps | 5 | The maximum queries per second (QPS) that can be performed on the Gimbal Kubernetes API server
| gimbal-client-burst | 10 | The maximum number of queries that can be performed on the Gimbal Kubernetes API server during a burst

### Credentials

The discoverer requires the following credentials to access the backend OpenStack cluster.
Similar to the OpenStack CLI, the credentials can be provided using environment variables:

| Credential         | Environment Variable  | Description                                       |
|--------------------|-----------------------|---------------------------------------------------|
| Username           | `OS_USERNAME`         | The OpenStack username                            |
| Password           | `OS_PASSWORD`         | The password of the OpenStack user                |
| Authentication URL | `OS_AUTH_URL`         | The URL of the endpoint to use for authentication |
| Tenant Name        | `OS_TENANT_NAME`      | The OpenStack user's tenant name                  |
| User Domain Name   | `OS_USER_DOMAIN_NAME` | The OpenStack user's domain name                  |

If you need to provide a CA certificate to establish a secure connection with the
authentication endpoint, you may use the `--openstack-certificate-authority` flag to
provide the path to a CA certificate.

#### Example

Following example creates a Kubernetes secret which the Openstack discoverer will consume to get credentials & other information to be able to discover services & endpoints:

```sh
kubectl create secret generic remote-discover-openstack \
    --from-file=certificate-authority-data=./ca.pem \
    --from-literal=backend-name=openstack \
    --from-literal=username=admin \
    --from-literal=password=abc123 \
    --from-literal=auth-url=https://api.openstack:5000/ \
    --from-literal=tenant-name=heptio
```

### Updating Credentials

Credentials to the backend OpenStack cluster can be updated at any time if necessary. To do so, we recommend taking advantage of the Kubernetes deployment's update features:

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

1. Connection is made to remote cluster and all LBaaS's and corresponding Members are retrieved from the cluster
2. Those objects are then translated into Kubernetes Services and Endpoints, then synchronized to the Gimbal cluster in the same namespace as the remote cluster. Labels will also be added during the synchronization (See the [labels](#labels) section for more details).
3. Once the initial list of objects is synchronized, any further updates will happen based upon the configured `reconciliation-period` which will start a new reconciliation loop.

### Labels

All synchronized services & endpoints will have additional labels added to assist in understanding where the object were sourced from. 

Labels added to service and endpoints:
```
gimbal.heptio.com/service=<serviceName>
gimbal.heptio.com/backend=<nodeName>
gimbal.heptio.com/load-balancer-id=<LoadBalancer.ID>
gimbal.heptio.com/load-balancer-name=<LoadBalancer..Name>
```