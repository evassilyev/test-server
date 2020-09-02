# test-server

The task have been described through the email
  
# Steps to run

## 1. prepare database
run database inside docker conainer and apply there scripts from the folder **db/scripts**:
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
*Note: **CGO_ENABLED=0** option required for run application inside Alpine Docker-image*

## 3. run server inside docker container
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
Application configured through setting environment variables listed below:  
  * SERVER_PORT - server's port. Example: 8803
  * SERVER_ENDPOINT - server's endpoint. Example: /endpoint
  * SERVER_PGDB - database connection URL. Example: postgres://enlabs:enlabs@localhost:5432/enlabs_test?sslmode=disable;
  * SERVER_PPINTERVAL - post processing start interval in minutes
