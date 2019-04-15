# auth-service

Authentication Service with ban option

## Run

Specify port on which to expose the auth-service in `docker-compose.override.yml`:

```yaml
version: '3'
services:
  redis:
    ports:
      - "6379:6379"
  auth-service:
    ports:
      - "8080:8080"
```

Start service with dependencies:
```sh
docker-compose up -d --build
```
---
## TODO
## Test

Run integration tests:
```sh
./integration.sh
```

## Build

Build auth-service:
```sh
go build ./cmd/auth-service
```

Run unit tests:
```sh
go test ./...
```

Run integration tests manually:
```sh
go test -tags integration -count 1 ./...
```
---