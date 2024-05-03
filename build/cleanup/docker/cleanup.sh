#!/bin/bash

NAMESPACES=("default" "murm-queue")

for NAMESPACE in "${NAMESPACES[@]}"
do
    echo "Checking pods in namespace: $NAMESPACE"
    kubectl get pods -n $NAMESPACE --no-headers | grep Terminating | while read pod_name; do
        echo "Deleting pod $pod_name stuck in terminating in $NAMESPACE..."
        kubectl delete pod $pod_name -n $NAMESPACE --force --grace-period=0
    done
done
