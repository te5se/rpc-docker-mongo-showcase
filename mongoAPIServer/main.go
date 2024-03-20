package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

var messages = []string{}

func main() {
	fmt.Println("started")
	go func() {
		kafkaReader, err := NewCatFactKafkaReader()
		if err != nil {
			log.Fatalf("error in creating kafka reader: %v", err)
		}

		partitionConsumer, err := kafkaReader.KafkaConsumer.ConsumePartition("cats", 0, sarama.OffsetNewest)
		if err != nil {
			log.Fatalf("error in creating kafka reader: %v", err)
		}
		for message := range partitionConsumer.Messages() {
			fmt.Printf("got message: %v\n", string(message.Value))
			messages = append(messages, string(message.Value))
		}
	}()

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, messages)
	})
	err := router.Run("localhost:8123")
	log.Fatalf("error in creating kafka reader: %v", err)
}

type CatFactKafkaReader struct {
	KafkaConsumer sarama.Consumer
}

func NewCatFactKafkaReader() (CatFactKafkaReader, error) {
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
	}, nil
}

func GetKafkaURI() string {
	var uri, exists = os.LookupEnv("KAFKA_URI")
	if exists {
		return uri
	}
	return "localhost:9092"
}
