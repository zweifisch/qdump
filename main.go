package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/streadway/amqp"
)

type Options struct {
	Amqpurl  string
	Topic    string
	Exchange string
	Queue    string
	Ttl      int32
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
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

	messages, err := channel.Consume(options.Queue, "qdump", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("basic.consume: %v", err)
	}

	go func() {
		for message := range messages {
			if message.ContentType == "application/json" {
				fmt.Println(string(message.Body))
			}
			message.Nack(false, true)
		}
	}()
}

func setup(options Options) {
	connection, err := amqp.Dial(options.Amqpurl)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}

	var args amqp.Table
	if options.Ttl > 0 {
		args = amqp.Table{"x-expires": options.Ttl}
	}

	queue := options.Queue
	if queue == "" {
		queue = options.Exchange + "-" + options.Topic + "-" + randSeq(4)
	}

	channel.QueueDeclare(queue, false, false, false, false, args)
	channel.QueueBind(queue, options.Topic, options.Exchange, false, nil)
	fmt.Printf("queue %s decleared\n", queue)
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
		cli.IntFlag{
			Name:  "ttl",
			Value: 0,
			Usage: "queue ttl",
		},
	}

	app.Action = func(c *cli.Context) {
		options := Options{
			Amqpurl:  c.String("amqpurl"),
			Exchange: c.String("exchange"),
			Topic:    c.String("topic"),
			Queue:    c.String("queue"),
			Ttl:      int32(c.Int("ttl")) * 1000,
		}
		if options.Topic != "" && options.Exchange != "" {
			setup(options)
		} else if options.Queue != "" {
			dump(options)
		}
	}

	app.Run(os.Args)
}
