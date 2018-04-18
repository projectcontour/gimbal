# Gimbal

Maintainers: [Heptio](https://github.com/heptio)

## Overview

Gimbal is a software defined service discovery and load balancing platform built on Kubernetes and Contour. It provides a scalable, multi-tenant, API-driven ingress tier that routes traffic to multiple endpoints. Contour functions as a Layer-7 software load balancer that can route traffic to backend Kubernetes and Openstack clusters based upon the configuration defined via Kubernetes Ingress resources. Services and endpoints within backend clusters are discovered via the Openstack & Kubernetes Discoverer components. 

![OverviewDiagram](docs/images/overview.png)

## Prerequisites

Gimbal is tested with Kubernetes clusters running version 1.9 and later but should work with any cluster starting at version 1.7.

## Get started

Deployment of Gimbal is outlined in the [deployment section](deployment/README.md) and also includes quick start applications.

## Documentation

Documentation on all the Gimbal components can be found on the [docs page](docs/README.md).

## Troubleshooting

If you encounter any problems that the documentation does not address, please [file an issue](https://github.com/heptio/gimbal/issues).

## Contributing

Thanks for taking the time to join our community and start contributing!

### Before you start

- Please familiarize yourself with the [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.
- See [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on the developer certificate of origin that we require.

### Pull Requests

- We welcome pull requests. Fee free to dig through the [issues](https://github.com/heptio/gimbal/issues) and jump in.
