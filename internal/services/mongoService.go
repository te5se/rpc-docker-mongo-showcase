package services

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoService struct {
	Connection *mongo.Client
}

func GetMongoService() (MongoService, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return MongoService{}, err
	}

	return MongoService{
		Connection: client,
	}, err
}

func (service MongoService) GetConnection() *mongo.Client {
	return service.Connection
}
