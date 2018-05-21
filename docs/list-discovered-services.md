# List Discovered Services

The Gimbal Discoverers add labels to the discovered services and endpoints before storing them in the Gimbal cluster. These labels are useful for querying the Gimbal cluster.

## List all discovered services and endpoints

```sh
kubectl get svc,endpoints -l gimbal.heptio.com/backend
```

You can add `--all-namespaces` to list across all namespaces in the Gimbal cluster.

## List services and endpoints that were discovered from a specific cluster

```sh
kubectl get svc,endpoints -l gimbal.heptio.com/backend=${CLUSTER_NAME}
```

## List services that belong to the same logical service

If you have instances of a service spread across clusters, you can use the `gimbal.heptio.com/service` label to list
them:

```sh
kubectl get svc,endpoints -l gimbal.heptio.com/service=${SERVICE_NAME}
```
