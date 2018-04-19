# Gimbal

Maintainers: [Heptio](https://github.com/heptio)

## Overview

Heptio Gimbal is a layer-7 load balancing platform built on Kubernetes, Envoy, and Contour. It provides a scalable, multi-team, and API-driven ingress tier capable of routing internet traffic to multiple upstream Kubernetes clusters and traditional infrastructure technologies including OpenStack.

Gimbal was developed out of a joint effort between Heptio and Yahoo Japan Corporation's subsidiary, Actapio, to modernize Yahoo! Japan’s infrastructure with Kubernetes without impacting legacy investments in OpenStack.

At launch, Gimbal can discover services from Kubernetes and OpenStack clusters, but we expect to support additional platforms in the future.

### Common Use Cases

* Organizations with multiple Kubernetes clusters that need a way to manage ingress traffic across them
* Organizations with Kubernetes and OpenStack infrastructure that need a consistent load balancing tier
* Organizations that want to enable their development teams to safely self-manage their routing configuration
* Organizations with bare metal or on-premises infrastructure that want cloud-like load balancing capabilities

![OverviewDiagram](docs/images/overview.png)

## Prerequisites

Gimbal is tested with Kubernetes clusters running version 1.9 and later but should work with any cluster starting at version 1.7.

Gimbal's service discovery is currently tested with Kubernetes 1.7+ and OpenStack Mitaka.

## Get started

Deployment of Gimbal is outlined in the [deployment section](deployment/README.md) and also includes quick start applications.

## Documentation

Documentation on all the Gimbal components can be found on the [docs page](docs/README.md).


## Known Limitations

* Upstream Kubernetes Pods and OpenStack VMs must be routable from the Gimbal load balancing cluster
  * No support for Kubernetes clusters with Overlay networks
  * We are looking for feedback on community requirements to design a solution. One potential option is to use one GRE tunnel per upstream cluster.  [Feedback welcome here](https://github.com/heptio/gimbal/issues/39)!
* Kubernetes Ingress API is limited and insecure
  * Only one backend per route
  * Anyone can modify route rules for a domain
  * More complex load balancing features like weighting and strategy are not supported
  * Gimbal & Contour will solve this with a [new IngressRoute CRD](https://github.com/heptio/contour/blob/master/design/ingressroute-design.md)

## Troubleshooting

If you encounter any problems that the documentation does not address, please [file an issue](https://github.com/heptio/gimbal/issues) or talk to us on the Kubernetes Slack team channel `#gimbal`.

## Contributing

Thanks for taking the time to join our community and start contributing!

Feedback and discussion is available on the [mailing list](https://groups.google.com/forum/#!forum/heptio-gimbal).

### Before you start

- Please familiarize yourself with the [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.
- See [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on the developer certificate of origin that we require.

### Pull Requests

- We welcome pull requests. Fee free to dig through the [issues](https://github.com/heptio/gimbal/issues) and jump in.
