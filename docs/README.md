# Gimbal

Gimbal is an overarching set of projects that includes Contour.  Gimbal builds on and expands to look at Load Balancing through more complicated environments.

Contour is concentrating on providing a solution for a single cluster.  Gimbal expands on this goal to solve scenarios that involve multiple clusters, and load balancing to non-Kubernetes resources.

## Overview Guides

The following guides will describe how components of Gimbal function and interact with other systems:

- [Kubernetes Discoverer](kubernetes-discoverer.md)
- [Openstack](openstack-discoverer.md)

Guides on how to setup / deploy Gimbal can be found in the [deployment guide](../deployment/README.md). 

## Operator Topics

- [Manage Backend Clusters and Discovery](manage-backends.md)
- [List Discovered Services](list-discovered-services.md)
- [Update Kubernetes Discoverer Credentials](kubernetes-discoverer.md#updating-credentials)
- [Update OpenStack Discoverer Credentials](openstack-discoverer.md#updating-credentials)

## User Topics

- [Route Specification](route.md)
- [Dashboards / Monitoring / Alerting](monitoring.md)
