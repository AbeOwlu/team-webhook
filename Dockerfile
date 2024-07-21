FROM golang:1.22-alpine AS builder

WORKDIR /webhook-server/

COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /cmd/webhook
RUN CGO_ENABLED=0 GOOS=linux go build -o main


FROM scratch
COPY --from=builder /webhook-server/cmd/webhook/main /

ENTRYPOINT [ "/main" ]