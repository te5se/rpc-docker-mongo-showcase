# rpc-docker-mongo-example

### This application is a showcase for
 - developing several Go applications from a single codebase
 - passing data by kafka messages
 - passing data by gRPC 
 - storing and retrieving data from mongoDB
 - running several apps in Docker from a single codebase
 - making microservices talk to each other and wait for initialization
 
### What does it do

- Scheduler service polls endpoint https://catfact.ninja/fact every 5 seconds and sends the payload to kafka topic
- MongoAPI service reads data from kafka topic and stores it in mongoDB
- Scheduler service provides endpoint http://localhost:9000. When visited, it requests data from scheduler service via gRPC and outputs it in JSON

### How to launch

- run docker-compose up
- visit localhost:9000
- use query params if you need to adjust page and pagesize of catfacts, for example, localhost:9000?page=0&size=1