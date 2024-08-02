package api

import (
	"Book-Library-Management-System/internal/models"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

var broadcast = make(chan models.Book)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
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
			client.Close()
			delete(clients, client)
		}
	}
}
