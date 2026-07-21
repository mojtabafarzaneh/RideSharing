package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	pbd "ride-sharing/shared/proto/driver"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DriverConsumer struct {
	rabbitMq *messaging.RabbitMQ
	service  domain.TripService
}

func NewDriverConsumer(rb *messaging.RabbitMQ, service domain.TripService) DriverConsumer {
	return DriverConsumer{
		rabbitMq: rb,
		service:  service,
	}
}

func (dc *DriverConsumer) Listen() error {
	return dc.rabbitMq.ConsumeMessages(messaging.DriverTripResponseQueue, func(ctx context.Context, msg amqp.Delivery) error {
		var driverEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &driverEvent); err != nil {
			log.Printf("failed to unmarshal message %s, err: %v", messaging.DriverTripResponseQueue, err)
			return err
		}
		var payload messaging.DriverTripResponseData
		if err := json.Unmarshal(driverEvent.Data, &payload); err != nil {
			log.Printf("failed to unmarshal message %s, err: %v", messaging.DriverTripResponseQueue, err)
			return err
		}
		switch msg.RoutingKey {
		case contracts.DriverCmdTripAccept:
			if err := dc.handleTripAccepted(ctx, payload.TripId, payload.Driver); err != nil {
				log.Printf("failed to handle the trip accept: %v", err)
				return err
			}
		case contracts.DriverCmdTripDecline:
			if err := dc.hanleTripDeclined(ctx, payload.TripId, payload.RiderId); err != nil {
				log.Printf("failed to handle the trip declined: %v", err)
				return err
			}
			return nil

		}

		return nil
	})
}

func (dc *DriverConsumer) handleTripAccepted(ctx context.Context, tripId string, driver *pbd.Driver) error {
	trip, err := dc.service.GetTripById(ctx, tripId)
	if err != nil {
		log.Printf("failed to get the trip: %v", err)
		return err
	}
	if trip == nil {
		log.Printf("couldn't find any trip with this id")
		return nil
	}

	if err := dc.service.UpdateTrip(ctx, tripId, "accepted", driver); err != nil {
		log.Printf("error occurred while updating trip: %v", err)
		return err
	}

	trip, err = dc.service.GetTripById(ctx, tripId)
	if err != nil {
		log.Printf("failed to get the trip: %v", err)
		return err
	}
	if trip == nil {
		log.Printf("couldn't find any trip with this id")
		return nil
	}

	marshaledTrip, err := json.Marshal(trip)
	if err != nil {
		log.Printf("couldn't marshal trip: %v", err)
		return err
	}

	if err := dc.rabbitMq.PublishMessage(ctx, contracts.TripEventDriverAssigned, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    marshaledTrip,
	}); err != nil {
		log.Printf("couldn't publish message %v", err)
		return err
	}

	marshalledPayload, err := json.Marshal(messaging.PaymentTripResponseData{
		TripID:   tripId,
		UserID:   trip.UserID,
		DriverID: driver.Id,
		Amount:   trip.RideFare.TotalPriceInCents,
		Currency: "USD",
	})

	if err := dc.rabbitMq.PublishMessage(ctx, contracts.PaymentCmdCreateSession,
		contracts.AmqpMessage{
			OwnerID: trip.UserID,
			Data:    marshalledPayload,
		},
	); err != nil {
		return err
	}
	return nil
}

func (dc *DriverConsumer) hanleTripDeclined(ctx context.Context, tripId string, riderId string) error {
	trip, err := dc.service.GetTripById(ctx, tripId)
	if err != nil {
		log.Printf("failed to get the trip: %v", err)
		return err
	}
	if trip == nil {
		log.Printf("couldn't find any trip with this id")
		return nil
	}

	newPayload := messaging.TripEventData{
		Trip: trip.ToPorto(),
	}

	marshaledPayload, err := json.Marshal(newPayload)
	if err != nil {
		log.Printf("couldn't marshal trip: %v", err)
		return err
	}

	if err := dc.rabbitMq.PublishMessage(ctx, contracts.TripEventDriverNotInterested, contracts.AmqpMessage{
		OwnerID: riderId,
		Data:    marshaledPayload,
	}); err != nil {
		log.Printf("couldn't publish message %v", err)
		return err
	}
	return nil
}
