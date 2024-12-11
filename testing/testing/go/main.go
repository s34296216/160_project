package main

// import {
// 	// "encoding/json"
// 	// "fmt"
// 	// "log"
// 	// "os"
// 	// "sync"
// 	// "net/http"
// 	"fmt"
// 	"log"

// 	"github.com/nats-io/nats.go"

// 	"github.com/nats-io/nats.go"
// 	zmq "github.com/pebbe/zmq4"
// 	"github.com/gorilla/websocket"
// }

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"

	zmq "github.com/pebbe/zmq4"
)

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
	Updates        int
	StockChanges   int
	LowStockAlerts int
	Events         map[int]int
	// ^^ each event for every product id
	// need mutex to keep locks and help threads
	mutex sync.Mutex
}

// making the new metrics for each event for each product
func NewMetric() *Metrics {
	return &Metrics{
		Events: make(map[int]int),
	}
}

// updating it with more messages
func (m *Metrics) UpdateMetrics(msg InventoryMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Updates++
	// debugging fixes this stockchange error
	// only stock change if there is changes happening
	if msg.Details.Change != nil {
		m.StockChanges += *msg.Details.Change
	}

	m.Events[msg.ProductID]++

	if msg.Status == "alert" {
		m.LowStockAlerts++
	}

}

// similar to the process of connecting to nats server
// process incoming data
func DataReceiver(address string, messageChannel chan<- InventoryMessage) {

	// example of SUB working
	// publisher, _ := zmq.NewSocket(zmq.PUB)
	// publisher.SetLinger(0)
	// defer publisher.Close()

	// publisher.Bind("tcp://127.0.0.1:9092")

	// subscriber, _ := zmq.NewSocket(zmq.SUB)
	// // subscriber.SetLinger(0)
	// defer subscriber.Close()

	// debugging, fixes declaring error
	context, err := zmq.NewContext()
	if err != nil {
		log.Fatal("Error creating ZMQ context:", err)
	}
	defer context.Term()

	publisher, _ := zmq.NewSocket(zmq.PUB)
	publisher.SetLinger(0)
	defer publisher.Close()

	publisher.Bind("tcp://127.0.0.1:9092")
	if err != nil {
		log.Fatal("Error creating ZMQ pub:", err)
	}

	// sub, err := context.NewSocket(zmq.sub)
	// sub, err := zmq.NewSocket(zmq.SUB)
	sub, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		log.Printf("Error creating ZMQ sub %v\n", err)
	}
	defer sub.Close()

	err = sub.Connect(address)
	if err != nil {
		log.Printf("Error connecting to ZeroMQ address %s: %v", address, err)
	}

	err = sub.SetSubscribe("")
	if err != nil {
		log.Printf("Error setting subscription: %v", err)
	}

	log.Printf("Connected to ZeroMQ at %s\n", address)

	// the above ensures a successful connection

	// start receiving messages
	// continuously
	for {
		mes, err := sub.Recv(0)
		if err != nil {
			log.Printf("Error receiving message %v\n", err)
			continue
		}

		var messages InventoryMessage
		if err := json.Unmarshal([]byte(mes), &messages); err != nil {
			log.Printf("Error parsing message: %v\n", err)
			continue
		}

		// sending the msg to the channel
		messageChannel <- messages

	}
}

// send results to connect clients in real time
type Broadcast struct {
	clients map[*websocket.Conn]bool
	mutex   sync.Mutex
}

func NewBroadcast() *Broadcast {
	return &Broadcast{
		clients: make(map[*websocket.Conn]bool),
	}
}

// new client add websocket
func (b *Broadcast) AddClient(conn *websocket.Conn) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.clients[conn] = true
}

// remove
func (b *Broadcast) RemoveClient(conn *websocket.Conn) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.clients, conn)
	conn.Close()
}

// broadcast msg
func (b *Broadcast) Broadcast(message interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for client := range b.clients {
		err := client.WriteJSON(message)
		if err != nil {
			log.Printf("Error sending message: %v", err)
			b.RemoveClient(client)
		}
	}
}

// upgrade
var upgradews = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// connects to websocket server
func WebSocketServer(b *Broadcast) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradews.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade error: %v", err)
			return
		}

		b.AddClient(conn)
		log.Println("Client added and connected")

		err = conn.WriteMessage(websocket.TextMessage, []byte("Hello, client!"))
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	})

	log.Println("WebSocket server started at ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (m *Metrics) Results() map[string]interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return map[string]interface{}{
		"total_updates":     m.Updates,
		"low_stock_alerts":  m.LowStockAlerts,
		"stock_changes":     m.StockChanges,
		"events_by_product": m.Events,
	}

}

func main() {
	// building bridge from NATS to ZeroMQ for communication between backend
	// Connect to NATS server
	nc, err := nats.Connect("nats://localhost:7921")
	if err != nil {
		log.Printf("Error connecting to NATS: %v\n", err)
		return
	}
	defer nc.Close()
	fmt.Println("Listening localhost:7921 ")

	// Create shared instances for metrics and broadcast module
	metrics := NewMetric()
	broadcast := NewBroadcast()

	// Channel for data received from the backend
	messageChannel := make(chan InventoryMessage)

	// Start WebSocket server
	go WebSocketServer(broadcast)

	// Start ZeroMQ Data Receiver
	address := "tcp://localhost:8888"
	go DataReceiver(address, messageChannel)

	nc.Subscribe("inventory.updates", func(msg *nats.Msg) {
		// part of the zeromq datareceiver function
		// needed in main to bridge nats to zeromq
		context, err := zmq.NewContext()
		if err != nil {
			log.Printf("Error creating ZMQ context: %v\n", err)
			return
		}
		defer context.Term()

		publisher, err := zmq.NewSocket(zmq.PUB)
		if err != nil {
			log.Printf("Error creating ZMQ pub: %v\n", err)
			return
		}
		defer publisher.Close()

		err = publisher.Bind("tcp://localhost:5555")
		if err != nil {
			log.Printf("Error binding ZMQ pub: %v\n", err)
			return
		}

		_, err = publisher.Send(string(msg.Data), 0)
		if err != nil {
			log.Printf("Error forwarding message ZMQ pub: %v\n", err)
			return
		} else {
			log.Printf("Success: forwarding message ZMQ pub: %s\n", string(msg.Data))
		}

	})
	// 	var messages []InventoryMessage
	// 	if err := json.Unmarshal(m.Data, &messages); err != nil {
	// 		fmt.Printf("Error parsing message: %v\n", err)
	// 		return
	// 	}

	// 	for _, msg := range messages {
	// 		fmt.Printf("Timestamp: %s\n", msg.Timestamp)
	// 		fmt.Printf("Product ID: %d\n", msg.ProductID)
	// 		fmt.Printf("Event: %s\n", msg.Event)
	// 		fmt.Printf("Status: %s\n", msg.Status)
	// 		fmt.Printf("Stock: %d\n", msg.Details.Stock)
	// 		if msg.Details.Change != nil {
	// 			fmt.Printf("Change: %d\n", *msg.Details.Change)
	// 		}
	// 		if msg.Details.Reason != nil {
	// 			fmt.Printf("Reason: %s\n", *msg.Details.Reason)
	// 		}
	// 		fmt.Printf("Message: %s\n", msg.Message)
	// 		fmt.Println("-------------------")
	// 	}
	// })
	// if err != nil {
	// 	fmt.Printf("Error subscribing: %v\n", err)
	// 	return
	// }
	// defer sub.Unsubscribe()

	// // Keep the connection alive
	// select {}

	// Main loop to process messages
	for msg := range messageChannel {
		// Update metrics
		metrics.UpdateMetrics(msg)

		// Broadcast updated metrics
		summary := metrics.Results()
		broadcast.Broadcast(summary)
	}

	// // Subscribe to inventory updates
	// sub, err := nc.Subscribe("inventory.updates", func(m *nats.Msg) {
	// 	var messages []InventoryMessage
	// 	if err := json.Unmarshal(m.Data, &messages); err != nil {
	// 		fmt.Printf("Error parsing message: %v\n", err)
	// 		return
	// 	}

	// 	for _, msg := range messages {
	// 		fmt.Printf("Timestamp: %s\n", msg.Timestamp)
	// 		fmt.Printf("Product ID: %d\n", msg.ProductID)
	// 		fmt.Printf("Event: %s\n", msg.Event)
	// 		fmt.Printf("Status: %s\n", msg.Status)
	// 		fmt.Printf("Stock: %d\n", msg.Details.Stock)
	// 		if msg.Details.Change != nil {
	// 			fmt.Printf("Change: %d\n", *msg.Details.Change)
	// 		}
	// 		if msg.Details.Reason != nil {
	// 			fmt.Printf("Reason: %s\n", *msg.Details.Reason)
	// 		}
	// 		fmt.Printf("Message: %s\n", msg.Message)
	// 		fmt.Println("-------------------")
	// 	}
	// })
	// if err != nil {
	// 	fmt.Printf("Error subscribing: %v\n", err)
	// 	return
	// }
	// defer sub.Unsubscribe()

	// // Keep the connection alive
	// select {}
}
