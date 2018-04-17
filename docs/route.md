# Route Specification

The core of Gimbal is Routes, which allow traffic into one or more applications. This section will discuss how to utilize Kubernetes `Ingress` objects to create these routes. 

Before beginning it is important to understand how service discovery functions within Gimbal. The Discoverer components should be deployed per upstream cluster. Once synchronized, services will show up in your team namespace with the cluster name appended.

For example, if a Kubernetes cluster is being discovered and there was a service named `s1` which existed in the namespace `team1`, in the cluster `node02`, once synchronized the service in the Gimbal cluster would be named `s1-node02`. 

## Basic Route

Following is a basic route which routes any request to `foo.bar.com` and proxies to a service named `s1` on the remote cluster `node02` over port `80`. 

```sh
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - backend:
          serviceName: s1-node02
          servicePort: 80
```

## Additional Information

More information regarding Ingress can be found here: [https://kubernetes.io/docs/concepts/services-networking/ingress/](https://kubernetes.io/docs/concepts/services-networking/ingress/)