package event

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setUp()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (Consumer *Consumer) setUp() error {
	channel, err := Consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declearExchange(channel)
}

type RPCPayload struct {
	Filename string `json:"filename"`
	Audio    string `json:"audio"` // Base64 encoded image
}
type Payload struct {
	Name string     `json:"name"`
	Data RPCPayload `json:"data"`
}

func (Consumer *Consumer) Listen(topics []string) error {
	ch, err := Consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declearQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		ch.QueueBind(
			q.Name,
			s,
			"process_audio",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message [Exchage, Queue] [process_audio, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "process":
		err := processAudio(payload)
		if err != nil {
			log.Print(err)
		}

	default:
		err := processAudio(payload)
		if err != nil {
			log.Print(err)
		}
	}
}

func processAudio(audio Payload) error {
	client, err := rpc.Dial("tcp", "localhost:5001")
	if err != nil {
		log.Print(err)
		return err
	}

	rpcPayload := RPCPayload{
		Filename: audio.Data.Filename,
		Audio:    audio.Data.Audio,
	}

	var result string
	err = client.Call("RPCServer.SaveAudio", rpcPayload, &result)
	if err != nil {
		return err
	}

	return nil
}
