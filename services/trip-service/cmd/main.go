package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/types"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9083"

func main() {

	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	rabbitmq, err := messaging.NewRabbitMQ("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("starting RabbitMQ Connection")

	publisher := events.NewTripEventPublisher(rabbitmq)

	grpcServer := grpcserver.NewServer()

	grpc.NewGRPCHandler(grpcServer, svc, publisher)

	log.Printf("gRPC server listening on %s", GrpcAddr)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()

	}()

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

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

type httpHandler struct {
	svc domain.TripService
}
