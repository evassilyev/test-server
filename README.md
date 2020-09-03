# test-server

The task has been described in the email
  
# Steps to run

## 1. prepare the database
run the database inside a docker container and apply there the scripts from the folder **db/scripts**:
```
docker run --name pgdbenlabs-test-task -e POSTGRES_USER=enlabs -e POSTGRES_PASSWORD=enlabs -p 54321:5432 -v $(pwd)/db/scripts/00_create_database.sql:/docker-entrypoint-initdb.d/init.sql -d postgres
cat db/scripts/01_create_schema.sql | docker exec -i pgdbenlabs-test-task psql -U enlabs -d enlabs_test 
```
if you want to add the additional negative balance check on the database side execute this:
```
cat db/scripts/02_create_additional_trigger_check.sql | docker exec -i pgdbenlabs-test-task psql -U enlabs -d enlabs_test
```
## 2. build the server
```
CGO_ENABLED=0 go build -o ./dist/app ./cmd/server  
```
*Note: **CGO_ENABLED=0** option is required to run the application inside Alpine Docker-image*

## 3. run the server inside a docker container
```
docker image build -f Dockerfile -t test-server .
docker run -p 8803:8803 \
--env SERVER_PORT=8803 \
--env SERVER_ENDPOINT=/endpoint \
--env SERVER_PGDB=postgres://enlabs:enlabs@localhost:54321/enlabs_test?sslmode=disable \
--env SERVER_PPINTERVAL=3 \
--net=host -d \
test-server:latest
```
# Or just run the bash script!
```
./run_everything.sh
```
# Additional info
## Configuration of the server
Application is configured through setting environment variables listed below:  
  * SERVER_PORT - server's port. Example: 8803
  * SERVER_ENDPOINT - server's endpoint. Example: /endpoint
  * SERVER_PGDB - database connection URL. Example: postgres://enlabs:enlabs@localhost:54321/enlabs_test?sslmode=disable;
  * SERVER_PPINTERVAL - post processing start interval in minutes

## Test commands
CURL command for tests:
```
curl --location --request POST 'localhost:8803/endpoint' \
--header 'Source-Type: server' \
--header 'Content-Type: application/json' \
--data-raw '{
    "state": "win",
    "amount": 20.00,
    "transactionId": "321110029000020017"
}'
```
Command for check the actual balance:
```
psql postgres://enlabs:enlabs@localhost:54321/enlabs_test?sslmode=disable -c 'select balance from calculated_balance_view'
```
Command for check the request history:
```
psql postgres://enlabs:enlabs@localhost:54321/enlabs_test?sslmode=disable -c 'select * from balance_history order by date_time desc'
```
