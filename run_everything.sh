#!/bin/sh
DBCONTAINER=pgdbenlabs-test-task
DBPORT=54321
SERVERIMAGE=serverenlabs-test-task
SERVER_PORT=8803
SERVER_ENDPOINT=/endpoint
SERVER_PGDB=postgres://enlabs:enlabs@${DBCONTAINER}:5432/enlabs_test?sslmode=disable
#SERVER_PGDB=postgres://enlabs:enlabs@$localhost:${DBPORT}/enlabs_test?sslmode=disable
SERVER_PPINTERVAL=2
SERVERCONTAINER=enlabs-test-server

echo "Initial parameters:"
echo "Database container name: ${DBCONTAINER}"
echo "Database container port: ${DBPORT}"
echo "Server image name: ${SERVERIMAGE}"
echo "Server container name: ${SERVERCONTAINER}"
echo "Server port: ${SERVER_PORT}"
echo "Server endpoint: ${SERVER_ENDPOINT}"
echo "Postgress connection URL: ${SERVER_PGDB}"
echo "Post process interval: ${SERVER_PPINTERVAL}"
echo ""


echo "Running database container and initializing the database..."
docker rm --force ${DBCONTAINER}
docker run --name ${DBCONTAINER} -e POSTGRES_USER=enlabs -e POSTGRES_PASSWORD=enlabs -p ${DBPORT}:5432 -v \
$(pwd)/db/scripts/00_create_database.sql:/docker-entrypoint-initdb.d/init.sql -d postgres

echo "Waiting while database waked up..."
sleep 5

echo "Creating tables & views..."
cat db/scripts/01_create_schema.sql | docker exec -i ${DBCONTAINER} psql -U enlabs -d enlabs_test

echo "Creating triggers..."
cat db/scripts/02_create_additional_trigger_check.sql | docker exec -i ${DBCONTAINER} psql -U enlabs -d enlabs_test

echo "Building the server..."
CGO_ENABLED=0 go build -o ./dist/app ./cmd/server

echo "Building image..."
docker image build -f Dockerfile -t ${SERVERIMAGE} .
echo "Running container..."
docker rm --force ${SERVERCONTAINER}
docker run --name ${SERVERCONTAINER} \
-p ${SERVER_PORT}:${SERVER_PORT} \
--env SERVER_PORT=${SERVER_PORT} \
--env SERVER_ENDPOINT=${SERVER_ENDPOINT} \
--env SERVER_PGDB=${SERVER_PGDB} \
--env SERVER_PPINTERVAL=${SERVER_PPINTERVAL} \
--link=${DBCONTAINER} -d \
${SERVERIMAGE}:latest
#--host=net

echo ""
echo "Server started on http://localhost:${SERVER_PORT}${SERVER_ENDPOINT}"
echo "Post process is running every ${SERVER_PPINTERVAL} minutes"




