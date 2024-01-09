package rmq

import (
	"context"
	"fmt"
	"net"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

type Exchange struct {
	Name string
	Kind string
}

type Queue struct {
	Name string
}

type RMQ struct {
	config    *Config
	conn      *amqp.Connection
	channel   *amqp.Channel
	exchanges []Exchange
	queues    []Queue
}

func New(config *Config, exchanges []Exchange, queues []Queue) *RMQ {
	return &RMQ{
		config:    config,
		exchanges: exchanges,
		queues:    queues,
	}
}

func (rmq *RMQ) Connect() error {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", rmq.config.Username, rmq.config.Password, net.JoinHostPort(rmq.config.Host, rmq.config.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to rmq on %s:%s: %w", rmq.config.Host, rmq.config.Port, err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open rmq channel: %w", err)
	}

	rmq.conn = conn
	rmq.channel = channel

	for _, exchange := range rmq.exchanges {
		err := rmq.DeclareExchange(exchange.Name, exchange.Kind)
		if err != nil {
			return err
		}
	}

	for _, queue := range rmq.queues {
		_, err := rmq.DeclareQueue(queue.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rmq *RMQ) DeclareExchange(exchangeName, kind string) error {
	err := rmq.channel.ExchangeDeclare(
		exchangeName,
		kind,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed do declare rmq exchange %q: %w", exchangeName, err)
	}

	return nil
}

func (rmq *RMQ) DeclareQueue(queueName string) (amqp.Queue, error) {
	var queue amqp.Queue

	queue, err := rmq.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return queue, fmt.Errorf("failed to declare rmq queue: %w", err)
	}

	return queue, nil
}

func (rmq *RMQ) Publish(ctx context.Context, exchangeName, routingKey string, message []byte) error {
	err := rmq.channel.PublishWithContext(ctx, exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "text/json",
		Body:        message,
	})
	if err != nil {
		return fmt.Errorf("failed to publish rmq message to exchange %q: %w", exchangeName, err)
	}

	return nil
}

func (rmq *RMQ) BindQueueToExchange(queueName, exchangeName, routingKey string) error {
	err := rmq.channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue %q to exchange %q: %w", queueName, exchangeName, err)
	}

	return nil
}

func (rmq *RMQ) Consume(_ context.Context, queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := rmq.channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages from queue %q: %w", queueName, err)
	}

	return msgs, nil
}

func (rmq *RMQ) Disconnect() error {
	if err := rmq.channel.Close(); err != nil {
		return fmt.Errorf("failed to close rmq channel: %w", err)
	}

	if err := rmq.conn.Close(); err != nil {
		return fmt.Errorf("failed to close rmq connection: %w", err)
	}

	return nil
}
