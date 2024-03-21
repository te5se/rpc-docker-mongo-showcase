# syntax=docker/dockerfile:1

# specify the base image to  be used for the application, alpine or ubuntu
FROM golang:1.22 as build

# create a working directory inside the image
WORKDIR /app

# copy Go modules and dependencies to image

COPY go.mod go.sum .
RUN go mod download

COPY . ./



RUN cd mongoAPIServer && go build -o /godocker
# tells Docker that the container listens on specified network ports at runtime

FROM ubuntu
COPY --from=build /godocker /godocker
COPY ./buildScripts/wait-for-it.sh /app/wait-for-it.sh

EXPOSE 8123

# command to be used to execute when the image is used to start a container
CMD ["/app/wait-for-it.sh", "kafka:9092", "--", "/godocker"]

#ENTRYPOINT ["tail"]
#CMD ["-f","/dev/null"]