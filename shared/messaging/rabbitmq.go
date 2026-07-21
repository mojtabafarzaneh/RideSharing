package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

const (
	TripExchange       = "trip"
	DeadLetterExchange = "dlx"
)

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to start a channel, %v", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	if err := rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to setup exchanges, %v", err)

	}
	return rmq, nil
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
	if r.Channel != nil {
		r.Channel.Close()
	}
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {
	marshaledMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.Channel.PublishWithContext(ctx,
		TripExchange, //exchange
		routingKey,   //routing key
		false,        //mandatory
		false,        //immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         marshaledMessage,
			DeliveryMode: amqp.Persistent,
		})

}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {

	err := r.Channel.Qos(
		1,     //prefetch count
		0,     //prefetch size
		false, //global: apply prefetchcount to each consumer individually
	)

	if err != nil {
		return fmt.Errorf("Failed to set Qos: %v", err)
	}
	msgs, err := r.Channel.Consume(
		queueName, //queue
		"",        //consumer
		false,     //auto-ack
		false,     //exclusive
		false,     //no-local
		false,     //no-wait
		nil,       //args
	)
	if err != nil {
		return err
	}

	ctx := context.Background()

	go func() {
		for msg := range msgs {
			var payload contracts.AmqpMessage
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				log.Printf("couldn't unmarshal message: %v", err)
			}
			log.Printf("Recived a message %s", payload)

			if err := handler(ctx, msg); err != nil {
				log.Printf("ERROR: Failed to handle message: %v. Message body: %s", err, msg.Body)
				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("ERROR: Failed to Nack message: %v", nackErr)
				}

				// Continue to the next message
				continue
			}

			// Only Ack if the handler succeeds
			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("ERROR: Failed to Ack message: %v. Message body: %s", ackErr, msg.Body)
			}
		}
	}()
	return nil
}
func (r *RabbitMQ) setupExchangesAndQueues() error {
	err := r.Channel.ExchangeDeclare(
		TripExchange, //name
		"topic",      // type
		true,         // durable
		false,        //auto-deleted
		false,        //internal
		false,        //no-wait
		nil,          //arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare exchange: %s : %v", TripExchange, err)
	}

	// args := amqp.Table{
	// 	"x-dead-letter-exchange": DeadLetterExchange,
	// }
	if err := r.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{
			contracts.TripEventCreated,
			contracts.TripEventDriverNotInterested,
		},
		TripExchange,
	); err != nil {
		return err

	}
	if err := r.declareAndBindQueue(
		DriverCMDTripRequestQueue,
		[]string{
			contracts.DriverCmdTripRequest,
		},
		TripExchange,
	); err != nil {
		return err

	}

	if err := r.declareAndBindQueue(
		DriverTripResponseQueue,
		[]string{contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverNoDriversFoundQueue,
		[]string{contracts.TripEventNoDriversFound},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		[]string{contracts.TripEventDriverAssigned},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSessionCreatedQueue,
		[]string{contracts.PaymentEventSessionCreated},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		PaymentTripResponseQueue,
		[]string{contracts.PaymentCmdCreateSession},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSessionCreatedQueue,
		[]string{contracts.PaymentEventSessionCreated},
		TripExchange,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string) error {
	q, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments with DLX config
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messageTypes {

		if err = r.Channel.QueueBind(
			q.Name,
			msg,
			exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue %s : %v", q.Name, err)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to queue bind %v : %s", TripExchange, err)
	}
	return nil
}
