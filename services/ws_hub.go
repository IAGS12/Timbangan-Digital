package services

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type WSHub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan interface{}
	mutex     sync.Mutex
}

var GlobalWSHub = NewWSHub()

func NewWSHub() *WSHub {
	hub := &WSHub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan interface{}, 100),
	}
	go hub.run()
	return hub
}

func (h *WSHub) Register(conn *websocket.Conn) {
	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()
	log.Println("[WS] Client WebSocket terhubung dari", conn.RemoteAddr())
}

func (h *WSHub) Unregister(conn *websocket.Conn) {
	h.mutex.Lock()
	if _, ok := h.clients[conn]; ok {
		delete(h.clients, conn)
		conn.Close()
		log.Println("[WS] Client WebSocket terputus:", conn.RemoteAddr())
	}
	h.mutex.Unlock()
}

func (h *WSHub) Broadcast(data interface{}) {
	h.broadcast <- data
}

func (h *WSHub) run() {
	for {
		msg := <-h.broadcast
		jsonBytes, err := json.Marshal(msg)
		if err != nil {
			log.Println("[WS] Error marshal broadcast message:", err)
			continue
		}

		h.mutex.Lock()
		for client := range h.clients {
			if err := client.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				log.Println("[WS] Error kirim pesan ke client:", err)
				client.Close()
				delete(h.clients, client)
			}
		}
		h.mutex.Unlock()
	}
}
