package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/te5se/rpc-docker-mongo-example/internal/scheduler"
	"github.com/te5se/rpc-docker-mongo-example/internal/services"
	"github.com/te5se/rpc-docker-mongo-example/pb"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/gin-gonic/gin"
)

func main() {
	var insecureClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	var mongoService, err = services.GetMongoService()
	if err != nil {
		log.Fatalf("error in mongo connection initialization: %v", err)
	}

	catFactKafkaWriter, err := NewCatFactKafkaWriter()
	if err != nil {
		log.Fatalf("error in kafka writer initialization: %v", err)
	}

	var catFactWorker = NewCatFactWorker(mongoService, catFactKafkaWriter, insecureClient)
	catFactWorker.start()

	router := gin.New()

	router.GET("/", getFacts)
	err = router.Run("0.0.0.0:8081")
	if err != nil {
		log.Fatalf("failed to run server:%v", err)
	}
}

func getFacts(ctx *gin.Context) {
	fmt.Println("got query")

	var page = ctx.Query("page")
	var size = ctx.Query("size")

	var intPage = 0
	var intSize = 0

	var err error

	if page != "" {
		intPage, err = strconv.Atoi(page)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, nil)
			return
		}
	}
	if size != "" {
		intSize, err = strconv.Atoi(size)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, nil)
			return
		}
	}

	if intSize == 0 {
		intSize = 20
	}

	client, err := scheduler.GetClient()
	if err != nil {
		log.Fatalf("couldn't get grpcClient connection: %v", err)
	}

	facts, err := client.GetFacts(context.TODO(), &pb.GetFactsRequest{
		PageSize: int64(intSize),
		Page:     int64(intPage),
	})
	if err != nil {
		fmt.Printf("get facts error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(200, facts)
}

type CatFactKafkaWriter struct {
	KafkaProducer sarama.SyncProducer
}

func NewCatFactKafkaWriter() (CatFactKafkaWriter, error) {
	var config = sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	kafkaProducer, err := sarama.NewSyncProducer([]string{GetKafkaURI()}, config)
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
	HTTPClient   *http.Client
}

func NewCatFactWorker(mongoService services.MongoService, kafkaWriter CatFactKafkaWriter, httpCient *http.Client) CatFactWorker {
	return CatFactWorker{
		MongoService: mongoService,
		KafkaWriter:  kafkaWriter,
		HTTPClient:   httpCient,
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

			resp, err := worker.HTTPClient.Get("https://catfact.ninja/fact")
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
func GetKafkaURI() string {
	var uri, exists = os.LookupEnv("KAFKA_URI")
	if exists {
		return uri
	}
	return "localhost:9092"
}
