FROM golang:1.22-alpine AS builder

WORKDIR /webhook-server

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.io,direct
RUN go mod download

COPY ./cmd/*/*.go ./internal/*/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main


FROM scratch
COPY --from=builder /webhook-server/main /

ENTRYPOINT [ "/main" ]