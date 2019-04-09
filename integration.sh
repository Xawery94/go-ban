#/bin/sh

docker-compose up --build -d

export TEST_ENDPOINT="$(docker-compose port auth-service 8080)"
export TEST_NATS="redis://$(docker-compose port redis 6379)"

go test -tags integration -count 1 -v ./...
EXIT_CODE=$?

docker-compose down -v

exit $EXIT_CODE
