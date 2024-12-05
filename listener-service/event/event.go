package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declearExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"process_audio",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

func declearQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
}
