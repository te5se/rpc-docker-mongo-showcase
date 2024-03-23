package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/IBM/sarama"
	mongoAPI "github.com/te5se/rpc-docker-mongo-example/internal/mongo-api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/te5se/rpc-docker-mongo-example/pb"
)

var messages = []string{}

func main() {
	fmt.Println("started")

	mongoConnection, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(GetMongoDBURI()))
	if err != nil {
		log.Fatalf("error in creating kafka reader: %v", err)
	}
	mongoCollection := mongoConnection.Database("catfactDB").Collection("catfacts")
	var mongoWriter = GetMongoWriter(mongoCollection)
	var catFactReader = GetMongoReader(mongoCollection)
	StartKafkaListener(mongoWriter)

	err = LaunchGRPCServer(catFactReader)
	if err != nil {
		log.Fatalf("server serve error: %v", err)
	}
}
func LaunchGRPCServer(reader CatFactMongoReader) error {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("failed to create listener: ", err)
	}
	s := grpc.NewServer()
	reflection.Register(s)

	pb.RegisterCatFactServiceServer(s, mongoAPI.GetGRPCServer(reader))
	if err := s.Serve(listener); err != nil {
		return err
	}

	return nil
}
func StartKafkaListener(mongoWriter CatFactMongoWriter) {
	go func() {
		kafkaReader, err := NewCatFactKafkaReader(mongoWriter)
		if err != nil {
			log.Fatalf("error in creating kafka reader: %v", err)
		}

		partitionConsumer, err := kafkaReader.KafkaConsumer.ConsumePartition("cats", 0, sarama.OffsetNewest)
		if err != nil {
			log.Fatalf("error in creating kafka reader: %v", err)
		}
		for message := range partitionConsumer.Messages() {
			err = mongoWriter.HandleMessage(message)
			if err != nil {
				fmt.Printf("write message error: %v\n", err)
			}
		}

	}()
}

type CatFactKafkaReader struct {
	KafkaConsumer sarama.Consumer
	MongoWriter   CatFactMongoWriter
}

func NewCatFactKafkaReader(mongoWriter CatFactMongoWriter) (CatFactKafkaReader, error) {
	var config = sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	var kafkaConsumer, err = sarama.NewConsumer([]string{GetKafkaURI()}, config)
	if err != nil {
		return CatFactKafkaReader{}, err
	}

	return CatFactKafkaReader{
		KafkaConsumer: kafkaConsumer,
		MongoWriter:   mongoWriter,
	}, nil
}

func GetKafkaURI() string {
	var uri, exists = os.LookupEnv("KAFKA_URI")
	if exists {
		return uri
	}
	return "localhost:9092"
}

type CatFactMongoWriter struct {
	Collection *mongo.Collection
}

func GetMongoWriter(connection *mongo.Collection) CatFactMongoWriter {
	return CatFactMongoWriter{
		Collection: connection,
	}
}

func (writer CatFactMongoWriter) HandleMessage(message *sarama.ConsumerMessage) error {
	fmt.Printf("got message: %v\n", string(message.Value))

	messages = append(messages, string(message.Value))

	// insert
	var bsonDocument = bson.M{}
	err := json.Unmarshal(message.Value, &bsonDocument)
	if err != nil {
		return err
	}

	factField, exists := bsonDocument["fact"]
	if !exists {
		return fmt.Errorf("fact field doesn't exists in document: %v", bsonDocument)
	}

	var factFilter = bson.D{{Key: "fact", Value: factField}}

	// check if exists
	existingValue := writer.Collection.FindOne(context.TODO(), factFilter)
	var returnedError = existingValue.Err() != nil && errors.Is(existingValue.Err(), mongo.ErrNoDocuments) == false
	var alreadyExists = existingValue.Err() == nil
	if alreadyExists {
		//return empty, already exists
		return nil
	} else if returnedError {
		return existingValue.Err()
	}

	// write in
	_, err = writer.Collection.InsertOne(context.TODO(), bsonDocument)
	if err != nil {
		return err
	}

	return nil
}

func GetMongoDBURI() string {
	var uri, exists = os.LookupEnv("MONGO_URI")
	if exists {
		return "mongodb://" + uri
	}
	return "mongodb://localhost:27017"
}

type CatFact struct {
	Fact   string `json:"fact"`
	Length string `json:"length"`
}

type CatFactMongoReader struct {
	Collection *mongo.Collection
}

func GetMongoReader(connection *mongo.Collection) CatFactMongoReader {
	return CatFactMongoReader{
		Collection: connection,
	}
}

func (reader CatFactMongoReader) GetCatFacts(page int, pageSize int) ([]string, error) {
	cursor, err := reader.Collection.Find(context.TODO(), bson.M{}, options.Find().SetProjection(bson.M{"fact": 1}).SetSkip(int64(pageSize)*int64(page)).SetLimit(int64(pageSize)))

	if err != nil {
		return []string{}, err
	}
	defer cursor.Close(context.TODO())

	var results = []CatFact{}
	/* var testResults = []bson.M{} */
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		return []string{}, err
	}

	var stringResults = make([]string, 0, len(results))
	for _, fact := range results {
		stringResults = append(stringResults, fact.Fact)
	}

	return stringResults, err
}
