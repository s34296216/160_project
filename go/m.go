package main

import {
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/nats-io/nats.go"
	zmq "github.com/pebbe/zmq4"
}

type InventoryDetails struct {
	Stock  int     `json:"stock_level"`
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

// aggregate metrics 
type Metrics struct {
	Updates   int
	StockChanges   int
	LowStockAlerts int
	Events map[int]int  
	// ^^ each event for every product id
}

// making the new metrics for each event for each product
func NewMetric() *Metrics {
	return &Metrics{
		Events: make(mape[int]int),
	}
}

//updating it with more messages
func UpdateMetrics (msg InventoryMessage) {
	msg.Updates++
	StockChanges += msg.Details.Change
	msg.Event[msg.ProductID]++
	msg.LowStockAlerts++

}

// similar to the process of connecting to nats server
func dataReceiver(address string) {
	sub, err := zmq4.NewSocket(zmq4.SUB)

	if err != nil {
		log.Printf("Error creating ZMQ %v\n", err)
	}

	defer sub.Close()

	if err := sub.Connect(address); err != nil {
		log.Printf("Error connecting to ZeroMQ address %s: %v", address, err)
	}

	if err := sub.SetSubscribe(""); err != nil {
		log.Printf("Error setting subscription: %v", err)
	}

	log.Printf("Connected to ZeroMQ at %s\n", address)

	// the above ensures a successful connection
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
