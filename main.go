package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/streadway/amqp"
)

type Options struct {
	Key      string
	Amqpurl  string
	Topic    string
	Exchange string
	Queue    string
}

func dump(options Options) {
	connection, err := amqp.Dial(options.Amqpurl)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}

	queue := options.Queue
	channel.QueueDeclare(queue, false, false, false, false, nil)
	channel.QueueBind(queue, options.Topic, options.Exchange, false, nil)

	messages, err := channel.Consume(queue, "qdump", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("basic.consume: %v", err)
	}

	go func() {
		for message := range messages {
			if message.ContentType == "application/json" {
				if options.Key == "" {
					fmt.Println(string(message.Body))
				} else {
					var body map[string]interface{}
					json.Unmarshal(message.Body, &body)
					fmt.Println(body[options.Key])
				}
			}
			message.Nack(false, true)
		}
	}()
}

func main() {

	app := cli.NewApp()
	app.Name = "qdump"
	app.Usage = "dump amqp queue"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "amqpurl",
			Value: "amqp://localhost",
			Usage: "amqp url",
		},
		cli.StringFlag{
			Name:  "key",
			Value: "",
			Usage: "key",
		},
		cli.StringFlag{
			Name:  "topic",
			Value: "",
			Usage: "topic",
		},
		cli.StringFlag{
			Name:  "exchange",
			Value: "",
			Usage: "exchange",
		},
		cli.StringFlag{
			Name:  "queue",
			Value: "",
			Usage: "queue",
		},
	}

	app.Action = func(c *cli.Context) {
		options := Options{
			Key:      c.String("key"),
			Amqpurl:  c.String("amqpurl"),
			Exchange: c.String("exchange"),
			Topic:    c.String("topic"),
			Queue:    c.String("queue"),
		}
		dump(options)
	}

	app.Run(os.Args)
}
