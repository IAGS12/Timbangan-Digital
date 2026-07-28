package services

import (
	"encoding/json"
	"fmt"

	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"smart-livestock-backend/models"
)

type MQTTService struct {
	client        mqtt.Client
	deviceService *DeviceService
	weightService *WeightService
	brokerURL     string
}

var GlobalMQTT *MQTTService

func NewMQTTService(brokerURL string, deviceService *DeviceService, weightService *WeightService) *MQTTService {
	if brokerURL == "" {
		brokerURL = "tcp://broker.emqx.io:1883"
	}

	svc := &MQTTService{
		brokerURL:     brokerURL,
		deviceService: deviceService,
		weightService: weightService,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(fmt.Sprintf("backend-smart-livestock-%d", time.Now().UnixNano()))
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(5 * time.Second)

	opts.OnConnect = func(c mqtt.Client) {
		log.Printf("✅ MQTT Client connected to broker: %s", brokerURL)
		svc.subscribeTopics()
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("⚠️ MQTT Connection lost: %v. Reconnecting...", err)
	}

	svc.client = mqtt.NewClient(opts)
	GlobalMQTT = svc
	return svc
}

func (s *MQTTService) Start() {
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("❌ Failed to connect to MQTT broker %s: %v", s.brokerURL, token.Error())
	}
}

func (s *MQTTService) Stop() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
		log.Println("🔌 MQTT Client disconnected.")
	}
}

func (s *MQTTService) subscribeTopics() {
	// 1. Subscribe to Live Weight & Telemetry: timbangan/+/live
	s.client.Subscribe("timbangan/+/live", 0, func(c mqtt.Client, m mqtt.Message) {
		var payload map[string]interface{}
		if err := json.Unmarshal(m.Payload(), &payload); err == nil {
			// Broadcast ke WebSocket frontend
			GlobalWSHub.Broadcast(payload)
		}
	})

	// 2. Subscribe to Weighing Final Record: timbangan/+/weighing
	s.client.Subscribe("timbangan/+/weighing", 1, func(c mqtt.Client, m mqtt.Message) {
		var req models.WeightRequest
		if err := json.Unmarshal(m.Payload(), &req); err == nil {
			log.Printf("📥 [MQTT WEIGHING] Terima data timbang dari %s: %.2f KG", req.DeviceID, req.Weight)
			_, err := s.weightService.AddWeight(req, nil)
			if err != nil {
				log.Printf("❌ [MQTT WEIGHING] Gagal simpan penimbangan: %v", err)
			}
		}
	})

	// 3. Subscribe to Pairing Request: timbangan/+/pair_req
	s.client.Subscribe("timbangan/+/pair_req", 1, func(c mqtt.Client, m mqtt.Message) {
		var req models.PairingRequest
		if err := json.Unmarshal(m.Payload(), &req); err == nil {
			log.Printf("📥 [MQTT PAIRING] Terima permintaan pairing dari device: %s", req.DeviceCode)
			err := s.deviceService.RequestPairing(req)
			if err != nil {
				log.Printf("⚠️ [MQTT PAIRING] Error request pairing: %v", err)
			}
			// Cek status pairing terkini di DB
			status, _ := s.deviceService.GetPairingStatus(req.DeviceCode)
			s.PublishPairingStatus(req.DeviceCode, status)
		}
	})
}

// PublishCommand mengirim perintah jarak jauh ke ESP32 via MQTT (misal: tare, select_cow, unpair)
func (s *MQTTService) PublishCommand(deviceCode string, action string, extraData map[string]interface{}) error {
	if s.client == nil || !s.client.IsConnected() {
		return fmt.Errorf("MQTT Client is not connected")
	}

	payload := map[string]interface{}{
		"action":      action,
		"device_code": deviceCode,
		"timestamp":   time.Now().Unix(),
	}

	for k, v := range extraData {
		payload[k] = v
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	topic := fmt.Sprintf("timbangan/%s/cmd", deviceCode)
	token := s.client.Publish(topic, 1, false, jsonBytes)
	token.Wait()

	log.Printf("📤 [MQTT CMD] Sent command '%s' to topic '%s'", action, topic)
	return token.Error()
}

// PublishPairingStatus mengirim update status pairing ke ESP32 (approved/pending/rejected/unpaired)
func (s *MQTTService) PublishPairingStatus(deviceCode string, status string) error {
	if s.client == nil || !s.client.IsConnected() {
		return fmt.Errorf("MQTT Client is not connected")
	}

	payload := map[string]interface{}{
		"pairing_status": status,
		"device_code":    deviceCode,
		"success":        true,
	}

	jsonBytes, _ := json.Marshal(payload)
	topic := fmt.Sprintf("timbangan/%s/pair_status", deviceCode)
	token := s.client.Publish(topic, 1, false, jsonBytes)
	token.Wait()

	log.Printf("📤 [MQTT PAIR STATUS] Sent status '%s' to topic '%s'", status, topic)
	return token.Error()
}
