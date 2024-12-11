package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/pebbe/zmq4"
)

func main() {
	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	fmt.Println("Connected to NATS server")

	// Setup ZeroMQ publisher
	zmqContext, err := zmq4.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	publisher, err := zmqContext.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	err = publisher.Bind("tcp://*:5555")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ZeroMQ publisher bound to port 5555")

	// Subscribe to NATS topic
	_, err = nc.Subscribe("your.topic", func(msg *nats.Msg) {
		_, err := publisher.Send(string(msg.Data), 0)
		if err != nil {
			log.Printf("Error forwarding message: %v", err)
			return
		}
		fmt.Printf("Forwarded message: %s\n", string(msg.Data))
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Subscribed to NATS topic: your.topic")

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigCh
	fmt.Println("\nShutting down...")
}
