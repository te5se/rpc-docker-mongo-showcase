package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/te5se/rpc-docker-mongo-example/internal/services"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	var mongoService, err = services.GetMongoService()
	if err != nil {
		log.Fatalf("error in mongo connection initialization: %v", err)
	}

	catFactKafkaWriter, err := NewCatFactKafkaWriter()
	if err != nil {
		log.Fatalf("error in kafka writer initialization: %v", err)
	}

	var catFactWorker = NewCatFactWorker(mongoService, catFactKafkaWriter)
	catFactWorker.start()

	<-make(chan int)
}

type CatFactKafkaWriter struct {
	KafkaProducer sarama.SyncProducer
}

func NewCatFactKafkaWriter() (CatFactKafkaWriter, error) {
	var config = sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	/* broker := sarama.NewBroker("localhost:9092")

	topic := "cat"
	topicDetail := &sarama.TopicDetail{}
	topicDetail.NumPartitions = int32(1)
	topicDetail.ReplicationFactor = int16(1)
	topicDetail.ConfigEntries = make(map[string]*string)

	topicDetails := make(map[string]*sarama.TopicDetail)
	topicDetails[topic] = topicDetail
	createTopicRequest := sarama.CreateTopicsRequest{
		Timeout:      time.Second * 15,
		TopicDetails: topicDetails,
	}
	err := broker.Open(config)
	if err != nil {
		return CatFactKafkaWriter{}, err
	}
	defer broker.Close()
	_, err = broker.CreateTopics(&createTopicRequest)
	if err != nil {
		return CatFactKafkaWriter{}, err
	} */

	kafkaProducer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		return CatFactKafkaWriter{}, err
	}

	return CatFactKafkaWriter{
		KafkaProducer: kafkaProducer,
	}, nil
}

type CatFactWorker struct {
	MongoService services.MongoService
	KafkaWriter  CatFactKafkaWriter
}

func NewCatFactWorker(mongoService services.MongoService, kafkaWriter CatFactKafkaWriter) CatFactWorker {
	return CatFactWorker{
		MongoService: mongoService,
		KafkaWriter:  kafkaWriter,
	}
}

func (kafkaWriter CatFactKafkaWriter) SendCat(cat bson.M) error {
	catJsonString, err := json.Marshal(cat)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{}
	msg.Topic = "cats"
	msg.Value = sarama.ByteEncoder(catJsonString)
	pid, offset, err := kafkaWriter.KafkaProducer.SendMessage(msg)

	if err != nil {
		return err
	}

	fmt.Printf("pid: %v, offset: %v\n", pid, offset)

	return nil
}

func (worker CatFactWorker) start() {
	go func() {
		for {
			time.Sleep(time.Second * 5)

			resp, err := http.Get("https://catfact.ninja/fact")
			if err != nil {
				log.Printf("catfact GET error: %v", err)
				continue
			}

			var catFact = bson.M{}

			err = json.NewDecoder(resp.Body).Decode(&catFact)
			if err != nil {
				log.Printf("catfact GET error: %v", err)
				continue
			}

			fmt.Printf("Catfact: %v\n", catFact)

			//send to kafka topic

			err = worker.KafkaWriter.SendCat(catFact)
			if err != nil {
				log.Printf("catfact GET error: %v", err)
				continue
			}
		}
	}()
}
