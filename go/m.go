package main

import {
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"net/http"

	"github.com/nats-io/nats.go"
	zmq "github.com/pebbe/zmq4"
	"github.com/gorilla/websocket"
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
	// need mutex to keep locks and help threads
	mutex sync.Mutex
}

// making the new metrics for each event for each product
func NewMetric() *Metrics {
	return &Metrics{
		Events: make(map[int]int),
	}
}

//updating it with more messages
func (m *Metrics) UpdateMetrics(msg InventoryMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Updates++
	m.StockChanges += msg.Details.Change
	m.Event[msg.ProductID]++

	if msg.Status == "alert" {
		m.LowStockAlerts++
	}

}

// similar to the process of connecting to nats server
// process incoming data
func DataReceiver(address string, messageChannel chan<- InventoryMessage) {

	sub, err := zmq.NewContext()
    if err != nil {
        log.Fatal("Error creating ZMQ context:", err)
    }
    defer sub.Term()

	sub, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatal("Error creating ZMQ sub %v\n", err)
	}
	defer sub.Close()

	err := sub.Connect(address)
	if err != nil {
		log.Fatal("Error connecting to ZeroMQ address %s: %v", address, err)
	}

	err := sub.SetSubscribe("")
	if err != nil {
		log.Fatal("Error setting subscription: %v", err)
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
			log.Printf("Error parsing message: %v\n", err, mes)
			continue
		}

		// sending the msg to the channel
		messageChannel <- messages


	}
}


// send results to connect clients in real time
type Broadcast struct {
	clients    map[*websocket.Conn]bool
	mutex         sync.Mutex
}

func NewBroadcast() *Broadcast {
	return &Broadcast{
		clients:    make(map[*websocket.Conn]bool),
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
var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// connects to websocket server
func WebSocketServer(b *Broadcast) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal("Upgrade error: %v", err)
			return
		}

		b.AddClient(conn)
		log.Println("Client added and conncected")
	})

	log.Println("WebSocket server started at ws://localhost:8080/ws")
	http.ListenAndServe(":8080", nil)
}

func (m *Metrics) Results() map[string]interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return map[string]interface{}{
		"total_updates":    m.Updates,
		"low_stock_alerts": m.LowStockAlerts,
		"stock_changes":    m.StockChanges,
		"events_by_product": m.Events,
	}
	
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
