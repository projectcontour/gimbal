# Managing Backends

## Add a new backend

To route traffic to a new backend, you must deploy a new Discoverer instance that discovers all Services and Endpoints and routes them appropriately.

### Kubernetes

1. Create a new Secret from the kubeconfig file for the cluster:

    ```sh
    BACKEND_NAME=new-k8s
    SECRET_NAME=${BACKEND_NAME}-discover-kubecfg
    kubectl -n gimbal-discovery create secret generic ${SECRET_NAME} \
        --from-file=config=./config \
        --from-literal=backend-name=${BACKEND_NAME}
    ```

1. Update the [deployment manfiest](../deployment/gimbal-discoverer/02-kubernetes-discoverer.yaml). Set the deployment name to the name of the new backend, and set the Secret name to the name of the new Secret.
1. Apply the updated manifest to the Gimbal cluster:

    ```sh
    kubectl -n gimbal-discovery apply -f new-k8s-discoverer.yaml
    ```

1. Verify the Discoverer is running by checking the number of available replicas in the new deployment, and by checking the logs of the new pod.

### OpenStack

1. Ensure you have all the required [credentials](./openstack-discoverer.md#credentials) for the OpenStack cluster.
1. Create a new Secret:

    ```sh
    BACKEND_NAME=new-openstack
    SECRET_NAME=${BACKEND_NAME}-discover-openstack
    kubectl -n gimbal-discovery create secret generic ${SECRET_NAME} \
        --from-file=certificate-authority-data=${CA_DATA_FILE} \
        --from-literal=backend-name=${BACKEND_NAME} \
        --from-literal=username=${OS_USERNAME} \
        --from-literal=password=${OS_PASSWORD} \
        --from-literal=auth-url=${OS_AUTH_URL} \
        --from-literal=tenant-name=${OS_TENANT_NAME}
    ```

1. Update the [deployment manifest](../deployment/gimbal-discoverer/02-openstack-discoverer.yaml). Set the deployment name to the name of the new backend, and update the secret name to the one created in the previous step.
1. Apply the updated manifest to the Gimbal cluster:

    ```sh
    kubectl -n gimbal-discovery apply -f new-openstack-discoverer.yaml
    ```

1. Verify the Discoverer is running by checking the number of available replicas in the new deployment, and by verifying the logs of the new pod.

## Remove a backend

To remove a backend from the Gimbal cluster, the Discoverer and the discovered services must be deleted.

### Delete the discoverer

1. Find the Discoverer instance that's responsible for the backend:

    ```sh
    # Assuming a Kubernetes backend
    kubectl -n gimbal-discovery get deployments -l app=kubernetes-discoverer
    ```

1. Delete the instance:

    ```sh
    kubectl -n gimbal-discovery delete deployment ${DISCOVERER_NAME}
    ```

1. Delete the Secret that holds the credentials for the backend cluster:

    ```sh
    kubectl -n gimbal-discovery delete secret ${DISCOVERER_SECRET_NAME}
    ```

### Delete Services and Endpoints

**Warning: Performing this operation results in Gimbal not sending traffic to this backend.**

1. List the Services that belong to the cluster, and verify the list:

    ```sh
    kubectl --all-namespaces get svc -l gimbal.heptio.com/backend=${CLUSTER_NAME}
    ```

1. List the namespaces with Services that were discovered:

    ```sh
    kubectl get svc --all-namespaces  -l gimbal.heptio.com/backend=${CLUSTER_NAME} -o jsonpath='{range .items[*]}{.metadata.namespace}{"\n"}{end}' | uniq
    ```

1. Iterate over the namespaces and delete all Services and Endpoints:

    ```sh
    NAMESPACES=$(kubectl get svc --all-namespaces  -l gimbal.heptio.com/backend=${CLUSTER_NAME} -o jsonpath='{range .items[*]}{.metadata.namespace}{"\n"}{end}' | uniq)
    for ns in $NAMESPACES
    do
        kubectl -n $ns delete svc,endpoints -l gimbal.heptio.com/backend=${CLUSTER_NAME}
    done
