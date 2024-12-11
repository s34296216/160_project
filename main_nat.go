package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to the NATS server
	nc, err := nats.Connect("nats://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	fmt.Println("Connected to NATS server")

	// Subscribe to the "logs" subject
	_, err = nc.Subscribe("logs", func(m *nats.Msg) {
		fmt.Printf("Received: %s\n", string(m.Data))
	})
	if err != nil {
		log.Fatal(err)
	}

	// Keep the connection alive
	select {}
}
