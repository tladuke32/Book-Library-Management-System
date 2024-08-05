package api

import (
	"Book-Library-Management-System/internal/models"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

var broadcast = make(chan models.Book)

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Client connected")
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("Received: %s", msg)
	}
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading to websocket: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer ws.Close()

	RegisterClient(ws)

	for {
		var book models.Book
		err := ws.ReadJSON(&book)
		if err != nil {
			log.Printf("error: %v", err)
			UnregisterClient(ws)
			break
		}
		broadcast <- book
	}
}

func HandleMessages() {
	for {
		book := <-broadcast
		NotifyClients(book)
	}
}

func RegisterClient(conn *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clients[conn] = true
}

func UnregisterClient(conn *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	delete(clients, conn)
}

func NotifyClients(book models.Book) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		err := client.WriteJSON(book)
		if err != nil {
			log.Printf("error writing JSON to client: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
