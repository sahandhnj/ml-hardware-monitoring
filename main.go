package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sahandhnj/ml-hardware-monitoring/db"
	"github.com/sahandhnj/ml-hardware-monitoring/gpu"
	"github.com/sahandhnj/ml-hardware-monitoring/types"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan types.Message)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()
	fmt.Println("HI")
	dbService, err := db.NewDBService()
	if err != nil {
		panic(err)
	}

	GPU := &gpu.GPU{
		DBService: dbService,
		Interval:  time.Second / 1000,
		Broadcast: broadcast,
	}

	go GPU.Run()

	log.Println("http server started on :8000")
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg types.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
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
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
