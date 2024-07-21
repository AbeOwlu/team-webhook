FROM alpine

WORKDIR /prerequisite/

COPY . .

RUN apk update && \
    apk add curl openssl && \
    chmod u+x deployreq.sh && \
    chmod u+x cert.sh && \
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mkdir -p ~/.local/bin && \
    mv ./kubectl ~/.local/bin/kubectl && \
    PATH=$PATH:$HOME/.local/bin

ENTRYPOINT [ "./deployreq.sh" ]