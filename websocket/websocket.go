package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	clients   map[*websocket.Conn]bool
	broadcast chan Message
}

type Message struct {
	Timestamp string `json:"time_stamp"`
	GPU       string `json:"gpu"`
}

func NewWebSocket() *WebSocket {
	return &WebSocket{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan Message),
	}
}

func (w *WebSocket) Start() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", w.handleConnections)

	log.Println("http server started on :9500")
	err := http.ListenAndServe(":9500", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (wss *WebSocket) handleConnections(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	wss.clients[ws] = true
}

func (wss *WebSocket) SendMessage(msg Message) {
	wss.broadcast <- msg
}
