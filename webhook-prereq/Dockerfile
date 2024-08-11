FROM alpine

ADD . ./prereq/

WORKDIR /prereq

RUN apk update && \
    apk add curl openssl && \
    chmod u+x deployreq.sh && \
    chmod u+x cert.sh && \
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv ./kubectl /usr/local/bin/kubectl

ENTRYPOINT [ "./deployreq.sh" ]