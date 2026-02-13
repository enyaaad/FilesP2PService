package handlers

import (
	"net/http"

	"github.com/backend-app/backend/internal/api/middleware"
	"github.com/backend-app/backend/internal/webrtc"
	"github.com/gin-gonic/gin"
)

type WebRTCHandler struct {
	turnServer *webrtc.Server
}

func NewWebRTCHandler(turnServer *webrtc.Server) *WebRTCHandler {
	return &WebRTCHandler{
		turnServer: turnServer,
	}
}

type TurnCredentialsResponse struct {
	TurnServers []string `json:"turn_servers"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Realm       string   `json:"realm"`
}

// GetTurnCredentials godoc
// @Summary Получение TURN credentials
// @Description Возвращает TURN серверы и credentials для WebRTC
// @Tags webrtc
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} TurnCredentialsResponse "TURN credentials"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Router /webrtc/turn-credentials [get]
func (h *WebRTCHandler) GetTurnCredentials(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	username, password := h.turnServer.GetTurnCredentials()
	turnServers := h.turnServer.GetTurnServers()

	c.JSON(http.StatusOK, TurnCredentialsResponse{
		TurnServers: turnServers,
		Username:    username,
		Password:    password,
		Realm:       h.turnServer.GetRealm(),
	})
}
