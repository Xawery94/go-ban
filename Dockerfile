FROM golang:1.11 AS build
WORKDIR /opt/auth-service

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=unspecified
RUN go install -ldflags="-X main.version=${VERSION}" ./cmd/auth-service

FROM gcr.io/distroless/base
WORKDIR /opt/auth-service
COPY --from=build /go/bin/auth-service ./

ENTRYPOINT [ "/opt/auth-service/auth-service" ]
