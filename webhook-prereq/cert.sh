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
namepsace=$2
secret=$3

csrname=${service}.${namepsace}
tmpdir=$(mktemp -d -q)

echo -e "creating certs in temp working dir ${tmpdir}"

touch ${tmpdir}/${service}.conf
cat <<EOF> ${tmpdir}/${service}.conf
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${service}
DNS.2 = ${service}.${namespace}
DNS.3 = ${service}.${namespace}.svc
EOF

# gen CA key
openssl genrsa -out ${tmpdir}/key.pem 2048
# create csr no sign
openssl req -new -key ${tmpdir}/key.pem -subj "/CN=${service}.${namespace}.svc" -out ${tmpdir}/${service}.csr -config ${tmpdir}/${service}.conf

kubectl delete csr ${csrname} 2>/dev/null || true

# https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/#create-certificatessigningrequest
cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: ${csrname}
spec:
  signerName: kubernetes.io/kubelet-serving
  request: $(cat ${tmpdir}/${service}.csr | base64 | tr -d "\n")
  usages:
  - server auth
  - digital signature
  - key encipherment
EOF

while true; do
    kubectl get csr ${csrname}
    if [ "$?" -eq 0 ]; then
        break
    fi
done

# approve csr for KAS to the webhook
kubectl certificate approve ${csrname}

# check certificate status for issued,approved update
for x in {1..9}; do
    serverCert=$(kubectl get csr ${csrname} -o jsonpath='{.status.certificate}') || true
    serverStatus=$(kubectl get csr ${csrname} -o jsonpath='{.status.conditions[].status}') || true
    if [[ ${serverCert} != '' ]]; then
        break
    fi
    sleep 1
done
if [[ ${serverCert} == '' ]] || [[ ${serverStatus} == "Denied" ]]; then
    echo "Certificate not issued. Giving up after 10 attempts. Manually check cert: ${csrname}" >&1
    kubectl certificate approve ${csrname}
    sleep 3
fi

echo ${serverCert} | openssl base64 -d -A -out ${tmpdir}/cert.pem

# create tls secret for webhook
kubectl create secret -n ${namepsace} ${secret} \
--cert=${tmpdir}/cert.pem \
--key=${tmpdir}/key.pem