# Documentation

Heptio Gimbal is a layer-7 load balancing platform built on Kubernetes, Envoy, and Contour. It provides a scalable, multi-team, and API-driven ingress tier capable of routing internet traffic to multiple upstream Kubernetes clusters and traditional infrastructure technologies including OpenStack.

## Data Flow

![Data Flow](images/data-flow.png)

## Architecture Overview

![Arch Overview](images/overview.png)

You can read more about the [Gimbal Architecture](gimbal-architecture.md).

## Overview Guides

The following guides will describe how components of Gimbal function and interact with other systems:
- [Kubernetes Discoverer](kubernetes-discoverer.md)
- [Openstack Discover](openstack-discoverer.md)

Guides on how to setup / deploy Gimbal can be found in the [deployment guide](../deployment/README.md). 

## Operator Topics

- [Manage Backend Clusters and Discovery](manage-backends.md)
- [List Discovered Services](list-discovered-services.md)
- [Update Kubernetes Discoverer Credentials](kubernetes-discoverer.md#updating-credentials)
- [Update OpenStack Discoverer Credentials](openstack-discoverer.md#updating-credentials)

## User Topics

- [Route Specification](route.md)
- [Dashboards / Monitoring / Alerting](monitoring.md)
