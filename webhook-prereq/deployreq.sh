#!/bin/bash

set -xe

Help(){
    echo  -e "Usage: cert.sh <kube-service> <service-namepsace> <secret-name>"
}

if [ "$#" -ne 3 ]; then
    Help
    exit
fi

service=$1
namespace=$2
secret=$3

#create namespace if not existing
kubectl create namespace ${namespace} || true

# Retrice cluster CA in configmap extension-apiserver-authentication; and generate tls secret
# CA_BUNDLE=`kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n'`
./cert.sh $service $namespace $secret

kubectl apply -f ./webhook.yaml