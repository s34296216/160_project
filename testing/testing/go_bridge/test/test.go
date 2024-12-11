package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pebbe/zmq4"
)

// Test message structure
type TestMessage struct {
	Timestamp      string `json:"timestamp"`
	UserIP         string `json:"user_ip"`
	Event          string `json:"event"`
	Endpoint       string `json:"endpoint"`
	ResponseTimeMS string `json:"response_time_ms"`
	Status         string `json:"status"`
}

func startPublisher() {
	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Sample messages
	messages := []string{
		`{"timestamp": "2024-11-20T10:15:31Z", "user_ip": "192.168.1.101", "event": "page_response", "endpoint": "/product", "response_time_ms": "1200", "status": "slow"}`,
		`{"timestamp": "2024-11-20T10:15:32Z", "user_ip": "192.168.1.102", "event": "page_response", "endpoint": "/checkout", "response_time_ms": "800", "status": "normal"}`,
	}

	// Publish messages every 2 seconds
	for i := 0; ; i++ {
		msg := messages[i%len(messages)]
		err := nc.Publish("your.topic", []byte(msg))
		if err != nil {
			log.Printf("Error publishing: %v", err)
			continue
		}
		fmt.Printf("Published message: %s\n", msg)
		time.Sleep(2 * time.Second)
	}
}

func startSubscriber() {
	// Setup ZeroMQ subscriber
	context, err := zmq4.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	subscriber, err := context.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatal(err)
	}
	defer subscriber.Close()

	err = subscriber.Connect("tcp://localhost:5555")
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe to all messages
	subscriber.SetSubscribe("")

	fmt.Println("ZeroMQ subscriber started, waiting for messages...")

	// Receive messages
	for {
		msg, err := subscriber.Recv(0)
		if err != nil {
			log.Printf("Error receiving: %v", err)
			continue
		}
		fmt.Printf("Received from ZeroMQ: %s\n", msg)
	}
}

func main() {
	// Start subscriber in a goroutine
	go startSubscriber()

	// Wait for subscriber to initialize
	time.Sleep(1 * time.Second)

	// Start publisher
	go startPublisher()

	// Keep program running
	select {}
}
