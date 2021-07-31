FROM golang:1.15.14-buster as builder

WORKDIR /go/src/app

# cache dependencies to accelerate most rebuilds
COPY go.mod go.sum ./
RUN go mod download all

COPY . .

RUN go fmt ./...
RUN go build -o /processor bin/processor/main.go

# processor image
FROM debian:buster-slim as processor
ADD https://curl.haxx.se/ca/cacert.pem /etc/pki/tls/certs/ca-bundle.crt

ENTRYPOINT ["./processor", "--config", "config.toml"]
COPY --from=builder /processor .
COPY config-processor.toml /config.toml