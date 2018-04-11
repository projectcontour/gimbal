# Gimbal

Maintainers: [Heptio](https://github.com/heptio)

## Overview

Gimbal is a software defined service discovery and load balancing platform built on Kubernetes and Contour. It provides a scalable, multi-tenant, API-driven ingress tier that routes traffic to multiple endpoints.

This guide will outline how to use Contour, the Kubernetes & Openstack Discoverers, and the Route [CRD (Custom Resource Definition)](https://kubernetes.io/docs/concepts/api-extension/custom-resources/) resource to define ingress routes that send traffic to discovered services. The Contour cluster functions as a Layer-7 software load balancer that can route traffic to multiple backend Kubernetes and Openstack clusters based upon the configuration defined in each Route CRD.

## Prerequisites

Gimbal is tested with Kubernetes clusters running version 1.9 and later.

## Get started

Deployment of Gimbal is outlined in the [deployment section](deployment/README.md). 

## Documentation

Guides on how the Discoverers work can be found here: 

- [Kubernetes](docs/discoverer/kubernetes/README.md)
- [Openstack](docs/discoverer/openstack/README.md)

Guides on how to setup metrics/monitoring can be found here:
- [Prometheus](deployment/prometheus/README.md)
- [Grafana](deployment/grafana/README.md)

## Troubleshooting

If you encounter any problems that the documentation does not address, please [file an issue](https://github.com/heptio/gimbal/issues).

## Contributing

Thanks for taking the time to join our community and start contributing!

### Before you start
- Please familiarize yourself with the [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.
- See [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on the developer certificate of origin that we require.

### Pull Requests
- We welcome pull requests. Fee free to dig through the [issues](https://github.com/heptio/gimbal/issues) and jump in.
