# Route Specification

The core of Gimbal is IngressRoutes, which allow traffic to be routed into one or more applications. This section will discuss how to utilize [Contour IngressRoute](https://github.com/heptio/contour/blob/master/docs/ingressroute.md) objects to create these definitions.

Before beginning it is important to understand how service discovery functions within Gimbal. The Discoverer components should be deployed per upstream cluster. Once synchronized, services will show up in your team namespace with the cluster name appended.

For example, if a Kubernetes cluster is being discovered and there was a service named `s1` which existed in the namespace `team1`, in the cluster `cluster1`, once synchronized, the service in the Gimbal cluster would be named `cluster1-s1` and it would be in the `team1` namespace.

## Basic Route

Following is a basic IngressRoute which routes any request to `foo.bar.com` and proxies to a service named `s1` on the remote cluster `node02` over port `80`.

```sh
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  name: test
spec:
  virtualhost:
    fqdn: foo.bar.com
  routes:
    - match: /
      services:
        - name: cluster1-service1
          port: 80
```

## IngressRoute Features

The IngressRoute API provides a number of [enhancements](https://github.com/heptio/contour/blob/master/docs/ingressroute.md#key-ingressroute-benefits) over Kubernetes Ingress:

* Weight shifting
* Multiple services per route
* Load balancing strategies
* Multi-team support

## IngressRoute Delegation

Gimbal's multi-team support is enabled through Contour's [IngressRoute Delegation](https://github.com/heptio/contour/blob/master/docs/ingressroute.md#ingressroute-delegation).

### Restricted root namespaces

Contour has an [enforcing mode](https://github.com/heptio/contour/blob/master/docs/ingressroute.md#restricted-root-namespaces) which accepts a list of namespaces where root IngressRoutes are valid.
Only users permitted to operate in those namespaces can therefore create IngressRoutes with the `virtualhost` field.

This restricted mode is enabled in Contour by specifying a command line flag, `--ingressroute-root-namespaces`, which will restrict Contour to only searching the defined namespaces for root IngressRoutes.

## Additional Information

More information regarding IngressRoutes can be found in the [Contour Documentation](https://github.com/heptio/contour/blob/master/docs/ingressroute.md)