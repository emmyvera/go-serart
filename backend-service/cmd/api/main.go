package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"serart_be/configuration"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	mongodbUrl = "mongodb://localhost:27017"
	WEB_PORT   = "80"
)

type Config struct {
	App    *configuration.Application
	Rabbit *amqp.Connection
}

func main() {
	// try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	app := Config{
		Rabbit: rabbitConn,
	}

	client, err := initMongoDB(mongodbUrl)
	if err != nil {
		log.Panic(err)
	}

	app.App = configuration.New(client)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", WEB_PORT),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panicf("error startin the server on port %s %v", WEB_PORT, err)
		return
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// don't continue until rabbit is ready

	for {
		c, err := amqp.Dial("amqp://guest:guest@localhost")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Connected to Rabitmq!")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
