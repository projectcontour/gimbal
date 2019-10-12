# <img src="docs/images/gimbal-logo.png" width="400px" > [![Build Status](https://travis-ci.org/vmware-tanzu/gimbal.svg?branch=master)](https://travis-ci.org/vmware-tanzu/gimbal)

## Overview

Gimbal is a layer-7 load balancing platform built on [Contour](https://vmware-tanzu.github.io/contour/), which is an Ingress controller for Kubernetes that works by deploying the [Envoy proxy](https://www.envoyproxy.io/) as a reverse proxy and load balancer. It provides a scalable, multi-team, and API-driven ingress tier capable of routing Internet traffic to multiple upstream Kubernetes clusters and to traditional infrastructure technologies such as OpenStack.

Gimbal was developed out of a joint effort between gimbal and Yahoo Japan Corporation's subsidiary, Actapio, to modernize Yahoo Japan’s infrastructure with Kubernetes, without affecting legacy investments in OpenStack.

Early releases of Gimbal can discover services that run on Kubernetes and OpenStack clusters, but support for additional platforms is expected in future releases.

### Common Use Cases

* Organizations with multiple Kubernetes clusters that need a way to manage ingress traffic across clusters
* Organizations with Kubernetes and OpenStack infrastructure that need a consistent load balancing tier
* Organizations that want to enable their development teams to safely self-manage their routing configuration
* Organizations with bare metal or on-premises infrastructure that want cloud-like load balancing capabilities

![OverviewDiagram](docs/images/overview.png)

## Supported versions

Gimbal runs on Kubernetes version 1.9 or later, but is tested to provide service discovery for clusters running Kubernetes 1.7 or later, or OpenStack Mitaka.

## Get started

Deployment of Gimbal is outlined in the [deployment section](deployment/README.md), which includes quick start applications.

## Documentation

Documentation for all the Gimbal components can be found in the [docs directory](docs/README.md).


## Known Limitations

* Upstream Kubernetes Pods and OpenStack VMs must be routable from the Gimbal load balancing cluster.
  * Support is not available for Kubernetes clusters with overlay networks.
  * We are looking for community feedback on design requirements for a solution. A possible option is one GRE tunnel per upstream cluster. [Feedback welcome here](https://github.com/vmware-tanzu/gimbal/issues/39)!

## Troubleshooting

If you encounter any problems that the documentation does not address, please [file an issue](https://github.com/vmware-tanzu/gimbal/issues) or talk to us on the Kubernetes Slack team channel [#gimbal](https://kubernetes.slack.com/messages/gimbal).

## Contributing

Thanks for taking the time to join our community and start contributing!

### Before you start

- Please familiarize yourself with the [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.
- See [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on the developer certificate of origin that we require.

### Pull Requests

- We welcome pull requests. Fee free to dig through the [issues](https://github.com/vmware-tanzu/gimbal/issues) and jump in.
