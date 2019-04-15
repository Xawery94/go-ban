FROM golang:1.11 AS build
WORKDIR /opt/auth-service

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV TEST_ENDPOINT=http://localhost:8080
CMD go test -count 1 -tags integration ./... -args -testEndpoint ${TEST_ENDPOINT}