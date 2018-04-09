# Gimbal Discoverer

[![Build Status](https://travis-ci.com/heptio/gimbal.svg?token=dGsEGqM7L7s2vaK7wDXC&branch=master)](https://travis-ci.com/heptio/gimbal)

## Overview
The Gimbal Discoverer currently has two different systems it can monitor, Kubernetes and Openstack. The purpose of the Discoverers are to perform service discovery for remote clusters by finding remote endpoints and synchronizing them to a central Kubernetes cluster as Services & Endpoints. 

### Kubernetes 
The Kubernetes discoverer monitors available Services and Endpoints for a single Kubernetes cluster. The credentials to access the each API server will be provided by the Administrators via a Kubernetes Secret.

The Discoverer will leverage the `watch` feature of the Kubernetes API to receive changes dynamically, rather than having to poll the API. All available services & endpoints will be synchronized to the Team namespace matching the source system.

### Openstack
The Openstack discoverer monitors all Load Balancer as a Service (LBaaS) configured as well as the corresponding Members. They are synchronized to the Team namespace as Services and Endpoints, with the Namespace being configured as the TenantName in Openstack. 

The Discoverer will poll the Openstack API on a customizable interval.  

## Get started

#### Args

Arguments are available to customize the discoverer:

| flag  | default  | description  | discoverer | 
|---|---|---|---|
| --version  |  false | Show version and quit  | ALL | 
| --num-threads  | 2  |  Specify number of threads to use when processing queue items. | ALL
| --gimbal-kubecfg-file  | ""  | Location of kubecfg file for access to kubernetes cluster hosting Gimbal | ALL
| --discover-kubecfg-file | ""  | Location of kubecfg file for access to remote kubernetes cluster to watch for services / endpoints | Kubernetes
| --cluster-name  | ""  |   Name of cluster scraping for services & endpoints | ALL
| --debug | false | Enable debug logging | ALL
| --reconciliation-period | 30s | The interval of time between reconciliation loop runs | Openstack
| --http-client-timeout | 5s | The HTTP client request timeout | Openstack
| --openstack-certificate-authority | "" | Path to cert file of the OpenStack API certificate authority | Openstack
| --resync-interval | 30m | Resync period for Kubernetes watch client | 

## Deployment

The discoverer can be deployed by utilizing the included deployment files. They contain the correct RBAC rules, as well as the deployment of the discoverer itself.

_NOTE: Best practice would be to to create a service account user in the remote cluster who only has permissions to `watch`, `list` and `get` services & endpoints._

### Kubernetes
```
# Create namespace / deployment / rbac rules:
$ kubectl apply -f deployment/kubernetes-discoverer

# Create secret for remote k8s cluster:
$ kubectl create secret generic remote-discover-kubecfg --from-file=./config -n gimbal-discoverer
```

### Openstack
```
# Create namespace / deployment / rbac rules:
$ kubectl apply -f deployment/openstack-discoverer

# Create secret for remote openstack cluster:
$ kubectl create secret generic remote-discover-openstack --from-literal=keystoneUrl=http://openstack001:5000/v3/ --from-literal=neutronUrl=http://openstack001:9696/ --from-literal=username=someUser --from-literal=password=secretPassword --from-literal=userDomain=default --from-file=./cert.pem -n gimbal-discoverer
```

## Development

### Kubernetes

The Kubernetes discoverer requires two configs, first is the Gimbal system which will run Contour and store services & endpoints, the other is the remote cluster to scrape for services & endpoints. The config file is standard kubeconfig file, just make sure it's named `config`. Please include any certs required to access to the remote cluster api:

```
$ go run cmd/kubernetes-discoverer/main.go --gimbal-kubecfg-file=./config --discover-kubecfg-file=./config --cluster-name=clustername
```

### Openstack

The Openstack discoverer requires the config for the Gimbal Kubernetes cluster which will run Contour and store services & endpoints, the other is the remote cluster to scrape for load balancers and members. The config file is standard kubeconfig file, just make sure it's named `config`. Please include any certs required to access to the remote cluster api:

```
$ OS_USERNAME=user OS_PASSWORD=password OS_AUTH_URL=https://url OS_TENANT_NAME=tenant go run cmd/openstack-discoverer/main.go --gimbal-kubecfg-file=./config --cluster-name=clustername
```

## Build / Test

```
Create a binary:
$ make build

Run tests:
$ make test

Create container:
$ REGISTRY=heptio make container

Push container: 
$ REGISTRY=heptio make push
```

_NOTE: The registry ENV variable allow you to override the registry so custom images can be tested._

## Contributing

Thanks for taking the time to join our community and start contributing!

#### Before you start

* Please familiarize yourself with the [Code of
Conduct](https://github.com/heptio/gimbal/blob/master/CODE_OF_CONDUCT.md) before contributing.
* See [CONTRIBUTING.md](https://github.com/heptio/gimbal/blob/master/CONTRIBUTING.md) for instructions on the
developer certificate of origin that we require.

#### Pull requests

* We welcome pull requests. Feel free to dig through the [issues](10) and jump in.