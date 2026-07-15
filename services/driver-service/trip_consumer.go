package main

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TripCosumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ, service *Service) *TripCosumer {
	return &TripCosumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *TripCosumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.FindAvailableDriversQueue, func(ctx context.Context, msg amqp.Delivery) error {
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			log.Printf("failed to unmarshal message %v", err)
			return err
		}
		var payload messaging.TripEventData

		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			log.Printf("failed to unmarshal message %v", err)
			return err
		}
		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return c.handleFindAndNotifyDrivers(ctx, payload)
		}

		log.Printf("driver received message %+v", payload)
		return nil
	})
}

func (c *TripCosumer) handleFindAndNotifyDrivers(ctx context.Context, payLoad messaging.TripEventData) error {
	suitable := c.service.FindAvailableDrivers(payLoad.Trip.SelectedFare.PackageSlug)

	log.Printf("found suitable driver %v", suitable)

	if len(suitable) == 0 {

		if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: payLoad.Trip.UserID,
		}); err != nil {
			return nil
		}
		return nil
	}

	suitableDriverId := suitable[0]

	marshaledEvent, err := json.Marshal(payLoad)
	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: suitableDriverId,
		Data:    marshaledEvent,
	}); err != nil {
		return nil
	}

	return nil
}
