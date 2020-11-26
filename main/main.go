package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// TODO: Come up with a less global structure?
var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel
var counts = make(chan Count)                // counts channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Count struct
type Count struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// Message struct
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Name string `json:"name"`
	Time string `json:"time"`
}

var templates = template.Must(template.ParseGlob("templates/*.html"))

func renderTemplate(w http.ResponseWriter, template string, data interface{}) {
	err := templates.ExecuteTemplate(w, template, data)
	if err != nil {
		internalServerError(w, err)
	}
}

func removeClient(client *websocket.Conn) {
	client.Close()
	delete(clients, client)
	counts <- Count{Type: "count", Count: len(clients)}
}

// / GET
func home(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true
	counts <- Count{Type: "count", Count: len(clients)}

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error reading message: %v", err)
			removeClient(ws)
			break
		}
		broadcast <- msg
	}
}

func handleCounts() {
	for {
		count := <-counts
		for client := range clients {
			err := client.WriteJSON(count)
			if err != nil {
				log.Printf("error writing count: %v", err)
				removeClient(client)
			}
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error writing message: %v", err)
				removeClient(client)
			}
		}
	}
}

func main() {
	// Init router
	r := mux.NewRouter()

	// Route handlers
	r.HandleFunc("/", home).Methods("GET")
	r.HandleFunc("/ws", handleConnections).Methods("GET")

	// Start listening for incoming counts and messages
	go handleCounts()
	go handleMessages()

	// Run server
	port := 8080
	fmt.Println(fmt.Sprintf("Serving on port %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
