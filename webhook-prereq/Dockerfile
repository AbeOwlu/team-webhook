FROM alpine

ADD . ./prereq/

WORKDIR /prereq

RUN apk update && \
    apk add curl openssl && \
    chmod u+x deployreq.sh && \
    chmod u+x cert.sh && \
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod u+x kubectl && \
    mkdir -p ~/.local/bin && \
    mv ./kubectl ~/.local/bin/kubectl && \
    PATH=$PATH:$HOME/.local/bin

CMD [ "./deployreq.sh" ]