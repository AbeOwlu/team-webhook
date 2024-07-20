#!/bin/bash

Usage() {
    echo -e "Usage: cert.sh <kube-service> <service-namepsace> <secret-name>"
}

if [ "$#" -ne 3 ]; then
    Usage
    exit
fi

