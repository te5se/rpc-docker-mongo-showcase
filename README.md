# rpc-docker-mongo-example

### This application is a showcase for
 - developing several Go applications from a single codebase
 - passing data by Kafka messages
 - passing data by gRPC 
 - storing and retrieving data from MongoDB
 - running several Go apps in Docker from a single codebase
 - making microservices talk to each other and wait for initialization
 - multi-stage docker builds
 
### What does it do

- Scheduler service polls endpoint https://catfact.ninja/fact every 5 seconds and sends the payload to Kafka topic
- MongoAPI service reads data from Kafka topic and stores it in MongoDB
- Scheduler service provides endpoint http://localhost:9000. When visited, it requests data from scheduler service via gRPC and outputs it in JSON

### How to launch

- run docker-compose up and wait for initialization
- visit http://localhost:9000
- use query params if you need to adjust page and pagesize of catfacts, for example, http://localhost:9000?page=0&size=1
