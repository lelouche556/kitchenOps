# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -p=1 -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -p=1 -o /out/topic-consumer ./cmd/topic-consumer
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -p=1 -o /out/temporal-worker ./cmd/temporal-worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -p=1 -o /out/kafka-tail ./cmd/kafka-tail

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/api /usr/local/bin/api
COPY --from=builder /out/topic-consumer /usr/local/bin/topic-consumer
COPY --from=builder /out/temporal-worker /usr/local/bin/temporal-worker
COPY --from=builder /out/kafka-tail /usr/local/bin/kafka-tail

EXPOSE 8080
CMD ["api"]
