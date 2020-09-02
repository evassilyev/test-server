!#/bin/bash
CGO_ENABLED=0 go build -o ./dist/app ./cmd/server
docker image build -f Dockerfile -t test-server .
docker run -p 8803:8803 \
--env SERVER_PORT=8803 \
--env SERVER_ENDPOINT=/endpoint \
--env SERVER_PGDB=postgres://enlabs:enlabs@localhost:5432/enlabs_test?sslmode=disable \
--env SERVER_PPINTERVAL=2 \
test-server:latest




