# Managing Backends

## Add a new backend

In order to route traffic to a new backend, you must deploy a new discoverer instance that will discover all the services and endpoints.

### Kubernetes

1. Obtain the cluster's kubeconfig file.
2. Create a new secret for the discoverer, using the kubeconfig obtained in the previous step:

    ```sh
    BACKEND_NAME=new-k8s
    SECRET_NAME=${BACKEND_NAME}-discover-kubecfg
    kubectl -n gimbal-discovery create secret generic ${SECRET_NAME} \
        --from-file=./kubeconfig \
        --from-literal=cluster-name=${BACKEND_NAME}
    ```

3. Update the [deployment manfiest](../deployment/gimbal-discoverer/02-kubernetes-discoverer.yaml). Set the deployment name to the name of the new backend, and update the secret name to the one created in the previous step.
4. Apply the updated manifest against the Gimbal cluster:

    ```sh
    kubectl -n gimbal-discovery apply -f new-k8s-discoverer.yaml
    ```
Your new backend Kubernetes cluster is now part of the Gimbal cluster.

### OpenStack

1. Ensure you have all the required [credentials](./openstack-discoverer.md#credentials) to the remote OpenStack cluster.
2. Create a new secret for the discoverer:

    ```sh
    BACKEND_NAME=new-openstack
    SECRET_NAME=${BACKEND_NAME}-discover-openstack
    kubectl -n gimbal-discovery create secret generic ${SECRET_NAME} \
        --from-file=certificate-authority-data=./ca.pem \
        --from-literal=cluster-name=${BACKEND_NAME} \
        --from-literal=username=admin \
        --from-literal=password=abc123 \
        --from-literal=auth-url=https://api.openstack:5000/ \
        --from-literal=tenant-name=heptio
    ```

3. Update the [deployment manifest](../deployment/gimbal-discoverer/02-openstack-discoverer.yaml). Set the deployment name to the name of the new backend, and update the secret name to the one created in the previous step.
4. Apply the updated manifest against the Gimbal cluster:

    ```sh
    kubectl -n gimbal-discovery apply -f new-openstack-discoverer.yaml
    ```

Your new backend OpenStack cluster is now part of the Gimbal cluster.

## Remove a backend

To remove a backend from the Gimbal cluster, the discoverer and the discovered services must be deleted.

### Delete the discoverer

1. Find the discoverer instance responsable of the backend:

    ```sh
    kubectl -n gimbal-discovery get deployments
    ```

2. Delete the discoverer instance responsable of the backend:

    ```sh
    kubectl -n gimbal-discovery delete deployment ${DISCOVERER_NAME}
    ```

### Delete all services/endpoints that were discovered

**Warning: Performing this operation will result in Gimbal not sending traffic to this backend.**

1. List services that belong to the cluster, and verify the list:

    ```sh
    kubectl --all-namespaces get svc -l gimbal.heptio.com/cluster=${CLUSTER_NAME}
    ```

2. Get a list of namespaces that have services discovered from this cluster:

    ```sh
    kubectl get svc --all-namespaces  -l gimbal.heptio.com/cluster=${CLUSTER_NAME} -o jsonpath='{range .items[*]}{.metadata.namespace}{"\n"}{end}' | uniq
    ```

3. Iterate over the namespaces and delete all services and endpoints discovered from this cluster:

    ```sh
    NAMESPACES=$(kubectl get svc --all-namespaces  -l gimbal.heptio.com/cluster=${CLUSTER_NAME} -o jsonpath='{range .items[*]}{.metadata.namespace}{"\n"}{end}' | uniq)
    for ns in $NAMESPACES
    do
        kubectl -n $ns delete svc,endpoints -l gimbal.heptio.com/cluster=${CLUSTER_NAME}
    done
