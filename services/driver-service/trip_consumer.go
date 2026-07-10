package main

import (
	"context"
	"log"
	"ride-sharing/shared/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TripCosumer struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ) *TripCosumer {
	return &TripCosumer{
		rabbitmq: rabbitmq,
	}
}

func (c *TripCosumer) Listen() error {
	return c.rabbitmq.ConsumeMessages("hello", func(ctx context.Context, msg amqp.Delivery) error {
		log.Printf("driver received message %v", msg)
		return nil
	})
}
