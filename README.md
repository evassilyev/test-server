# test-server
Test task 

##build command  
```
CGO_ENABLED=0 go build -o ./dist/app ./cmd/server  
```
Note: CGO_ENABLED=0 required for run application inside Alpine Docker-image

##configuration
Application should be configured through setting environment variables listed below:  
  * SERVER_PORT - server's port. Example: 8803
  * SERVER_ENDPOINT - server's endpoint. Example: /endpoint
  * SERVER_PGDB - database connection URL. Example: postgres://enlabs:enlabs@localhost:5432/enlabs_test?sslmode=disable;
  * SERVER_PPINTERVAL - post processing start interval in minutes
  
##run server inside docker container
```
docker image build -f Dockerfile -t test-server .
docker run -p 8803:8803 \
--env SERVER_PORT=8803 \
--env SERVER_ENDPOINT=/endpoint \
--env SERVER_PGDB=postgres://enlabs:enlabs@localhost:5432/enlabs_test?sslmode=disable \
--env SERVER_PPINTERVAL=2 \
--net=host -d \
test-server:latest
```




