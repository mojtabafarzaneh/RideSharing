package messaging

import (
	"context"
	"fmt"
	"log"

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

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message string) error {
	return r.Channel.PublishWithContext(ctx,
		"",      //exchange
		"hello", //routing key
		false,   //mandatory
		false,   //immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(message),
			DeliveryMode: amqp.Persistent,
		})

}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	msgs, err := r.Channel.Consume(
		queueName, //queue
		"",        //consumer
		true,      //auto-ack
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
			log.Printf("Recived a message %s", msg.Body)

			if err := handler(ctx, msg); err != nil {
				log.Fatalf("Failed to handle the message %v", err)
			}
		}

	}()

	return nil
}
func (r *RabbitMQ) setupExchangesAndQueues() error {
	args := amqp.Table{
		"x-dead-letter-exchange": DeadLetterExchange,
	}

	_, err := r.Channel.QueueDeclare(
		"hello", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		args,    // arguments with DLX config
	)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
