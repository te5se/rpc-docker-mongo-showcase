package scheduler

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/te5se/rpc-docker-mongo-example/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var conn *grpc.ClientConn

func GetClient() (pb.CatFactServiceClient, error) {
	if conn == nil {
		serverAddr := flag.String(
			"server", GetServerURI(),
			"The server adress in the format of host:port",
		)

		flag.Parse()

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

		connLocal, err := grpc.DialContext(ctx, *serverAddr, opts...)
		if err != nil {
			log.Fatalln("failed dialing: ", err)
		}
		conn = connLocal
	}

	client := pb.NewCatFactServiceClient(conn)

	return client, nil
}
func GetServerURI() string {
	var uri, exists = os.LookupEnv("MONGO_API_URI")
	if exists {
		return uri
	}
	return "localhost:8080"
}
