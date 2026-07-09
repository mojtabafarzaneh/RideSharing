package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	lis, err := net.Listen("tcp", GrpcAddr)

	service := newService()
	grpcServer := grpcserver.NewServer()

	NewGrpcHandler(grpcServer, service)

	log.Printf("gRPC server listening on %s", GrpcAddr)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()

	}()
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to sreve: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Printf("shuting down the server")

	grpcServer.GracefulStop()
}

// type previewTripRequest struct {
// 	UserID      string           `json:"userID"`
// 	Pickup      types.Coordinate `json:"pickup"`
// 	Destination types.Coordinate `json:"destination"`
// }

// type httpHandler struct {
// 	svc domain.TripService
// }
