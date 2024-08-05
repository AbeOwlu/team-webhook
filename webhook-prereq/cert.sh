#!/bin/sh

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
openssl genrsa -out ${tmpdir}/ca.key 2048
openssl req -new -x509 -days 3650 -nodes -key ${tmpdir}/ca.key -out ${tmpdir}/ca.crt -subj "/CN=admissiona_cert_ca"

# gen rsa and sign with CA above
openssl genrsa -out ${tmpdir}/key.pem 2048
# create csr and sign
openssl req -new -key ${tmpdir}/key.pem -subj "/CN=${service}.${namespace}.svc" -out ${tmpdir}/${service}.csr -config ${tmpdir}/${service}.conf
openssl x509 -req -days 365 -in ${tmpdir}/${service}.csr -CA ${tmpdir}/ca.crt -CAkey ${tmpdir}/ca.key -CAcreateserial -out ${tmpdir}/server.pem -extensions v3_req -extfile ${tmpdir}/${service}.conf
# openssl x509 -req -in ${tmpdir}/${service}.csr -CA  ${tmpdir}/ca.crt -CAkey ${tmpdir}/ca.key -CAcreateserial -out ${tmpdir}/server.pem -days 365 -extensions v3_req -extfile csr.conf

# create tls secret for webhook
kubectl delete secret ${secret} -n ${namepsace} || true
sleep 1

kubectl create secret tls ${secret} -n ${namepsace} \
--cert=${tmpdir}/server.pem \
--key=${tmpdir}/key.pem

# # create the caBundle as a secret
# kubectl delete secret cabundle -n ${namepsace} || true
# sleep 1

# kubeclt create secret generic cabundle -n ${namepsace} \
# --from-file=cabundle=${tmpdir}/ca.crt 

export CA_BUNDLE=$(cat ${tmpdir}/ca.crt | base64 | tr -d '\n')

update webhook deployment file client service CA_BUNDLE
sed -i "s/CA_BUNDLE/$CA_BUNDLE/g" ./webhook.yaml


####------------------------------------------------------------------#####

## Breaking change with kubernets apiVersion <1.22

####------------------------------------------------------------------#####


# everything here is deprecated after kubernetes v1.22 - user cert-manager or manual cert injection
# kubectl delete csr ${csrname} 2>/dev/null || true

# # https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/#create-certificatessigningrequest
# cat <<EOF | kubectl apply -f -
# apiVersion: certificates.k8s.io/v1
# kind: CertificateSigningRequest
# metadata:
#   name: ${csrname}
# spec:
#   signerName: kubernetes.io/kubelet-serving
#   request: $(cat ${tmpdir}/${service}.csr | base64 | tr -d "\n")
#   usages:
#   - server auth
#   - digital signature
#   - key encipherment
# EOF

# while true; do
#     kubectl get csr ${csrname}
#     if [ "$?" -eq 0 ]; then
#         break
#     fi
# done

# # approve csr for KAS to the webhook
# kubectl certificate approve ${csrname}

# # check certificate status for issued,approved update
# for x in {1..9}; do
#     serverCert=$(kubectl get csr ${csrname} -o jsonpath='{.status.certificate}') || true
#     serverStatus=$(kubectl get csr ${csrname} -o jsonpath='{.status.conditions[].status}') || true
#     if [[ ${serverCert} != '' ]]; then
#         break
#     fi
#     sleep 1
# done
# if [[ ${serverCert} == '' ]] || [[ ${serverStatus} == "Denied" ]]; then
#     echo "Certificate not issued. Giving up after 10 attempts. Manually check cert: ${csrname}" >&1
#     kubectl certificate approve ${csrname}
#     sleep 3
#     serverCert=$(kubectl get csr ${csrname} -o jsonpath='{.status.certificate}') || true
# fi

# echo ${serverCert} | openssl base64 -d -A -out ${tmpdir}/cert.pem
