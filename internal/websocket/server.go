package websocket

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/backend-app/backend/internal/repository"
	"github.com/backend-app/backend/pkg/config"
	"github.com/backend-app/backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// SignalingMessage представляет сообщение для WebRTC signaling
type SignalingMessage struct {
	Type         string          `json:"type"` // "offer", "answer", "ice-candidate", "error"
	FromDeviceID string          `json:"from_device_id,omitempty"`
	ToDeviceID   string          `json:"to_device_id,omitempty"`
	SDP          *SDPMessage     `json:"sdp,omitempty"`
	Candidate    *ICECandidate   `json:"candidate,omitempty"`
	Error        string          `json:"error,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`
}

// SDPMessage содержит SDP offer или answer
type SDPMessage struct {
	Type string `json:"type"` // "offer" или "answer"
	SDP  string `json:"sdp"`
}

// ICECandidate содержит ICE candidate
type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMLineIndex int    `json:"sdpMLineIndex,omitempty"`
	SDPMid        string `json:"sdpMid,omitempty"`
}

// Client представляет подключенное устройство
type Client struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	DeviceID uuid.UUID
	Conn     *websocket.Conn
	Send     chan SignalingMessage
	Hub      *Hub
	LastSeen time.Time
	mu       sync.Mutex
}

// Hub управляет всеми подключенными клиентами
type Hub struct {
	clients    map[uuid.UUID]*Client // device_id -> client
	broadcast  chan SignalingMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	deviceRepo *repository.DeviceRepo
	log        zerolog.Logger
}

// Server представляет WebSocket сервер для signaling
type Server struct {
	hub    *Hub
	config *config.Config
	port   string
}

// NewServer создает новый WebSocket signaling сервер
func NewServer(cfg *config.Config, db *sql.DB, deviceRepo *repository.DeviceRepo) *Server {
	hub := &Hub{
		clients:    make(map[uuid.UUID]*Client),
		broadcast:  make(chan SignalingMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		deviceRepo: deviceRepo,
		log:        logger.Get(),
	}

	go hub.run()

	return &Server{
		hub:    hub,
		config: cfg,
		port:   cfg.Server.WebSocketPort,
	}
}

// Start запускает WebSocket сервер
func (s *Server) Start() error {
	http.HandleFunc("/ws/signaling", s.handleWebSocket)
	addr := fmt.Sprintf(":%s", s.port)
	s.hub.log.Info().Str("port", s.port).Msg("WebSocket signaling server starting")
	return http.ListenAndServe(addr, nil)
}

// handleWebSocket обрабатывает WebSocket подключения
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.hub.log.Error().Err(err).Msg("Failed to upgrade connection")
		return
	}

	deviceIDStr := r.URL.Query().Get("device_id")
	deviceToken := r.URL.Query().Get("device_token")

	if deviceIDStr == "" || deviceToken == "" {
		conn.WriteJSON(SignalingMessage{
			Type:  "error",
			Error: "device_id and device_token are required",
		})
		conn.Close()
		return
	}

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		conn.WriteJSON(SignalingMessage{
			Type:  "error",
			Error: "invalid device_id",
		})
		conn.Close()
		return
	}

	device, err := s.hub.deviceRepo.GetByToken(deviceToken)
	if err != nil || device == nil {
		conn.WriteJSON(SignalingMessage{
			Type:  "error",
			Error: "invalid device_token",
		})
		conn.Close()
		return
	}

	if device.ID != deviceID {
		conn.WriteJSON(SignalingMessage{
			Type:  "error",
			Error: "device_id does not match device_token",
		})
		conn.Close()
		return
	}

	s.hub.deviceRepo.UpdateLastSeen(deviceID)

	client := &Client{
		ID:       uuid.New(),
		UserID:   device.UserID,
		DeviceID: deviceID,
		Conn:     conn,
		Send:     make(chan SignalingMessage, 256),
		Hub:      s.hub,
		LastSeen: time.Now(),
	}

	s.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// run обрабатывает регистрацию/отмену регистрации клиентов и рассылку сообщений
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.DeviceID] = client
			h.mu.Unlock()
			h.log.Info().
				Str("device_id", client.DeviceID.String()).
				Msg("Client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.DeviceID]; ok {
				delete(h.clients, client.DeviceID)
				close(client.Send)
			}
			h.mu.Unlock()
			h.log.Info().
				Str("device_id", client.DeviceID.String()).
				Msg("Client unregistered")

		case message := <-h.broadcast:
			h.mu.RLock()
			if toDeviceID, err := uuid.Parse(message.ToDeviceID); err == nil {
				if client, ok := h.clients[toDeviceID]; ok {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(h.clients, toDeviceID)
					}
				} else {
					h.log.Warn().
						Str("device_id", message.ToDeviceID).
						Msg("Target device not connected")
				}
			}
			h.mu.RUnlock()
		}
	}
}

// readPump читает сообщения из WebSocket соединения
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg SignalingMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.log.Error().Err(err).Msg("WebSocket error")
			}
			break
		}

		c.mu.Lock()
		c.LastSeen = time.Now()
		c.mu.Unlock()

		c.handleMessage(msg)
	}
}

// writePump отправляет сообщения в WebSocket соединение
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				c.Hub.log.Error().Err(err).Msg("Failed to write message")
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage обрабатывает входящее сообщение
func (c *Client) handleMessage(msg SignalingMessage) {
	switch msg.Type {
	case "offer":
		c.handleOffer(msg)
	case "answer":
		c.handleAnswer(msg)
	case "ice-candidate":
		c.handleICECandidate(msg)
	default:
		c.Hub.log.Warn().
			Str("type", msg.Type).
			Msg("Unknown message type")
	}
}

// handleOffer обрабатывает SDP offer
func (c *Client) handleOffer(msg SignalingMessage) {
	if msg.ToDeviceID == "" {
		c.sendError("to_device_id is required for offer")
		return
	}

	toDeviceID, err := uuid.Parse(msg.ToDeviceID)
	if err != nil {
		c.sendError("invalid to_device_id")
		return
	}

	toDevice, err := c.Hub.deviceRepo.GetByID(toDeviceID)
	if err != nil || toDevice == nil {
		c.sendError("target device not found")
		return
	}

	if toDevice.UserID != c.UserID {
		c.sendError("devices must belong to the same user")
		return
	}

	c.Hub.broadcast <- SignalingMessage{
		Type:         "offer",
		FromDeviceID: c.DeviceID.String(),
		ToDeviceID:   msg.ToDeviceID,
		SDP:          msg.SDP,
	}
}

// handleAnswer обрабатывает SDP answer
func (c *Client) handleAnswer(msg SignalingMessage) {
	if msg.ToDeviceID == "" {
		c.sendError("to_device_id is required for answer")
		return
	}

	c.Hub.broadcast <- SignalingMessage{
		Type:         "answer",
		FromDeviceID: c.DeviceID.String(),
		ToDeviceID:   msg.ToDeviceID,
		SDP:          msg.SDP,
	}
}

// handleICECandidate обрабатывает ICE candidate
func (c *Client) handleICECandidate(msg SignalingMessage) {
	if msg.ToDeviceID == "" {
		c.sendError("to_device_id is required for ice-candidate")
		return
	}

	c.Hub.broadcast <- SignalingMessage{
		Type:         "ice-candidate",
		FromDeviceID: c.DeviceID.String(),
		ToDeviceID:   msg.ToDeviceID,
		Candidate:    msg.Candidate,
	}
}

// sendError отправляет сообщение об ошибке клиенту
func (c *Client) sendError(errorMsg string) {
	select {
	case c.Send <- SignalingMessage{
		Type:  "error",
		Error: errorMsg,
	}:
	default:
		c.Hub.log.Warn().Msg("Failed to send error message")
	}
}

// Shutdown останавливает WebSocket сервер
func (s *Server) Shutdown(ctx context.Context) error {
	s.hub.mu.Lock()
	for _, client := range s.hub.clients {
		client.Conn.Close()
	}
	s.hub.mu.Unlock()
	return nil
}
