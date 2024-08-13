FROM golang:1.22-alpine AS builder

WORKDIR /webhook-server

COPY . ./
ENV GOPROXY=https://goproxy.io,direct
ENV GOPRIVATE=github.com/AbeOwlu/team-webhook/*

RUN go mod download && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build ./cmd/webhook/main.go


FROM alpine
COPY --from=builder /webhook-server/main /

ENTRYPOINT [ "/main" ]