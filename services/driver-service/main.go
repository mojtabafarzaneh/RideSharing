package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/shared/messaging"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	service := newService()

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

	consumer := NewTripConsumer(rabbitmq, service)
	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("couldn't consume rabbitmq %v", err)
		}

	}()

	grpcServer := grpcserver.NewServer()

	NewGrpcHandler(grpcServer, service)
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
