package main

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type InventoryDetails struct {
	Stock  int     `json:"stock"`
	Change *int    `json:"change"`
	Reason *string `json:"reason"`
}

type InventoryMessage struct {
	Timestamp string           `json:"timestamp"`
	ProductID int              `json:"product_id"`
	Event     string           `json:"event"`
	Details   InventoryDetails `json:"details"`
	Status    string           `json:"status"`
	Message   string           `json:"message"`
}

func main() {
	// Connect to NATS server
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		fmt.Printf("Error connecting to NATS: %v\n", err)
		return
	}
	defer nc.Drain()
	fmt.Println("Listening localhost:4222 ")

	// Subscribe to inventory updates
	sub, err := nc.Subscribe("inventory.updates", func(m *nats.Msg) {
		var messages []InventoryMessage
		if err := json.Unmarshal(m.Data, &messages); err != nil {
			fmt.Printf("Error parsing message: %v\n", err)
			return
		}

		for _, msg := range messages {
			fmt.Printf("Timestamp: %s\n", msg.Timestamp)
			fmt.Printf("Product ID: %d\n", msg.ProductID)
			fmt.Printf("Event: %s\n", msg.Event)
			fmt.Printf("Status: %s\n", msg.Status)
			fmt.Printf("Stock: %d\n", msg.Details.Stock)
			if msg.Details.Change != nil {
				fmt.Printf("Change: %d\n", *msg.Details.Change)
			}
			if msg.Details.Reason != nil {
				fmt.Printf("Reason: %s\n", *msg.Details.Reason)
			}
			fmt.Printf("Message: %s\n", msg.Message)
			fmt.Println("-------------------")
		}
	})
	if err != nil {
		fmt.Printf("Error subscribing: %v\n", err)
		return
	}
	defer sub.Unsubscribe()

	// Keep the connection alive
	select {}
}
