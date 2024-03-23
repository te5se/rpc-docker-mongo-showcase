package grpcServer

import (
	"context"
	"fmt"

	"github.com/te5se/rpc-docker-mongo-example/internal/definitions"
	"github.com/te5se/rpc-docker-mongo-example/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedCatFactServiceServer
	CatFactReader definitions.CatFactReader
}

func (s *grpcServer) GetFacts(
	ctx context.Context, in *pb.GetFactsRequest,
) (*pb.GetFactsResponse, error) {

	facts, err := s.CatFactReader.GetCatFacts(int(in.Page), int(in.PageSize))
	if err != nil {
		fmt.Printf("get facts error: %v\n", err)
		return &pb.GetFactsResponse{
			Facts: []string{},
		}, status.Error(codes.Internal, "")
	}

	return &pb.GetFactsResponse{
		Facts: facts,
	}, nil
}

func GetGRPCServer(reader definitions.CatFactReader) *grpcServer {
	return &grpcServer{
		CatFactReader: reader,
	}
}
