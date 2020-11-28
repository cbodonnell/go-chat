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

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message struct
type Message struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}

// Count struct
type Count struct {
	Count int `json:"count"`
}

// Chat struct
type Chat struct {
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

// / GET
func home(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

// / GET
func handleConnections(w http.ResponseWriter, r *http.Request) {
	client, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	clients[client] = true
	broadcast <- Message{Type: "count", Body: Count{Count: len(clients)}}

	for {
		var msg Message
		err := client.ReadJSON(&msg)
		if err != nil {
			log.Printf("error reading message: %v", err)
			delete(clients, client)
			broadcast <- Message{Type: "count", Body: Count{Count: len(clients)}}
			break
		}
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error writing message: %v", err)
				client.Close()
				delete(clients, client)
				broadcast <- Message{Type: "count", Body: Count{Count: len(clients)}}
			}
		}
	}
}

func main() {
	// Init router
	r := mux.NewRouter()

	// Route handlers
	r.HandleFunc("/", home).Methods("GET")
	r.HandleFunc("/chat", handleConnections).Methods("GET")

	// Static file handler
	r.PathPrefix("/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	// Start listening for incoming messages
	go handleMessages()

	// Run server
	port := 8080
	fmt.Println(fmt.Sprintf("Serving on port %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
