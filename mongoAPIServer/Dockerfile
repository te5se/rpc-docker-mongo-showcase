# syntax=docker/dockerfile:1

# specify the base image to  be used for the application, alpine or ubuntu
FROM golang:1.22

# create a working directory inside the image
WORKDIR /app

# copy Go modules and dependencies to image
COPY . ./


# download Go modules and dependencies
run go mod vendor

# copy directory files i.e all files ending with .go
# compile application

RUN cd mongoAPIServer && go build -mod vendor -o /godocker

COPY ./buildScripts/wait-for-it.sh /app/wait-for-it.sh

# tells Docker that the container listens on specified network ports at runtime
EXPOSE 8123

# command to be used to execute when the image is used to start a container
CMD ["/app/wait-for-it.sh", "kafka:9092", "--", "/godocker"]