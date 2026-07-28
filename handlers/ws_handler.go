package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"smart-livestock-backend/services"
)

// UpgradeToWebSocket Middleware untuk memverifikasi upgrade koneksi WebSocket
func UpgradeToWebSocket(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// HandleWebSocket Handler utama koneksi WebSocket
func HandleWebSocket(c *websocket.Conn) {
	services.GlobalWSHub.Register(c)
	defer services.GlobalWSHub.Unregister(c)

	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		// Echo balik pesan ping jika ada
		if string(msg) == "ping" {
			_ = c.WriteMessage(mt, []byte("pong"))
			continue
		}

		// Parse JSON message dari client/ESP32 dan broadcast ke semua listener WebSocket
		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err == nil {
			services.GlobalWSHub.Broadcast(payload)
		}
	}
}
