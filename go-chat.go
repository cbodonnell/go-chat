package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

// --- Configuration --- //

var config Configuration

func getConfig(ENV string) Configuration {
	file, err := os.Open(fmt.Sprintf("config.%s.json", ENV))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var config Configuration
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

// TODO: Come up with a less global structure?
var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Configuration struct
type Configuration struct {
	Debug   bool   `json:"debug"`
	Port    int    `json:"port"`
	SSLCert string `json:"sslCert"`
	SSLKey  string `json:"sslKey"`
	JWTKey  string `json:"jwtKey"`
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

// Group struct
type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// var templates = template.Must(template.ParseGlob("templates/*.html"))

// func renderTemplate(w http.ResponseWriter, template string, data interface{}) {
// 	err := templates.ExecuteTemplate(w, template, data)
// 	if err != nil {
// 		internalServerError(w, err)
// 	}
// }

// / GET
// func home(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "index.html", nil)
// }

// / GET
func handleConnections(w http.ResponseWriter, r *http.Request) {
	client, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// badRequest(w, err)
		fmt.Println(err)
		return
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

func jwtMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := checkClaims(r)
		if err != nil {
			fmt.Println("Unauthorized request from " + r.RemoteAddr + " - " + err.Error())
			unauthorizedRequest(w, err)
			return
		}
		fmt.Println(claims)
		h.ServeHTTP(w, r)
	})
}

func main() {
	// Get configuration
	ENV := os.Getenv("ENV")
	if ENV == "" {
		ENV = "dev"
	}
	fmt.Println(fmt.Sprintf("Running in ENV: %s", ENV))
	config = getConfig(ENV)

	// Start listening for incoming messages
	go handleMessages()

	// Init router
	r := mux.NewRouter()

	// Route handlers
	// r.HandleFunc("/", home).Methods("GET")
	r.HandleFunc("/chat", handleConnections).Methods("GET")

	// CORS in dev environment
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowCredentials: true,
	})

	// Run server
	port := config.Port
	fmt.Println(fmt.Sprintf("Serving on port %d", port))

	if ENV == "dev" {
		r.Use(cors.Handler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	}

	r.Use(jwtMiddleware)
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), config.SSLCert, config.SSLKey, r))
}
