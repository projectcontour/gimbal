#!/bin/bash

NUM_SERVICES=$1

if [[ -z "$NUM_SERVICES" ]]
then
    echo "Must provide expected number of services"
    exit 1
fi

kubectl -n gimbal-discovery scale deploy openstack-discoverer --replicas=1

START=$(date +%s)

discovered=0
while [[ "$discovered" != "$NUM_SERVICES" ]]
do
    echo "total discovered $discovered"
    sleep 1
    discovered=$(kubectl get svc --all-namespaces -l gimbal.heptio.com/backend=openstack -o go-template='{{len .items}}')
done


END=$(date +%s)
echo "Full discovery of $NUM_SERVICES took $((END-START)) seconds"

exit 0

kubectl -n gimbal-discovery scale deploy openstack-discoverer --replicas=0

sleep 1

kubectl get svc --all-namespaces -l gimbal.heptio.com/backend=openstack \
    -o jsonpath='{range .items[*]}-n {.metadata.namespace} {.metadata.name}{"\n"}{end}' | \
    xargs -L 1 kubectl delete svc