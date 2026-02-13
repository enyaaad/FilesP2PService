package webrtc

import (
	"context"
	"fmt"
	"net"

	"github.com/backend-app/backend/pkg/config"
	"github.com/backend-app/backend/pkg/logger"
	"github.com/pion/turn/v3"
	"github.com/rs/zerolog"
)

// Server представляет TURN сервер для ретрансляции WebRTC трафика
type Server struct {
	server   *turn.Server
	config   *config.WebRTCConfig
	log      zerolog.Logger
	listener net.PacketConn
}

// NewTurnServer создает новый TURN сервер
func NewTurnServer(cfg *config.WebRTCConfig) (*Server, error) {
	log := logger.Get()

	publicIP := cfg.TURNPublicIP
	if publicIP == "" {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err == nil {
			localAddr := conn.LocalAddr().(*net.UDPAddr)
			publicIP = localAddr.IP.String()
			conn.Close()
			log.Info().Str("ip", publicIP).Msg("Auto-detected local IP address")
		} else {
			publicIP = "127.0.0.1"
			log.Warn().Msg("TURN_PUBLIC_IP not set, using 127.0.0.1 (for development only)")
		}
	}

	listener, err := net.ListenPacket("udp4", fmt.Sprintf("0.0.0.0:%s", cfg.TURNPort))
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP listener: %w", err)
	}

	relayIP := net.ParseIP(publicIP)
	if relayIP == nil {
		return nil, fmt.Errorf("invalid relay IP address: %s", publicIP)
	}

	listeningAddr := listener.LocalAddr().(*net.UDPAddr)

	relayGen := &turn.RelayAddressGeneratorPortRange{
		RelayAddress: relayIP,
		MinPort:      49152,
		MaxPort:      65535,
		MaxRetries:   10,
		Address:      fmt.Sprintf("%s:0", listeningAddr.IP.String()),
	}

	s, err := turn.NewServer(turn.ServerConfig{
		Realm:       cfg.TURNRealm,
		AuthHandler: createAuthHandler(cfg.TURNUsername, cfg.TURNPassword),
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn:            listener,
				RelayAddressGenerator: relayGen,
			},
		},
	})
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to create TURN server: %w", err)
	}

	log.Info().
		Str("port", cfg.TURNPort).
		Str("realm", cfg.TURNRealm).
		Str("public_ip", publicIP).
		Msg("TURN server created")

	return &Server{
		server:   s,
		config:   cfg,
		log:      log,
		listener: listener,
	}, nil
}

// createAuthHandler создает обработчик аутентификации для TURN
func createAuthHandler(username, password string) turn.AuthHandler {
	return func(usernameFromClient, realm string, srcAddr net.Addr) ([]byte, bool) {
		if usernameFromClient == username {
			key := turn.GenerateAuthKey(username, realm, password)
			return key, true
		}
		return nil, false
	}
}

// Start запускает TURN сервер (блокирующий вызов)
func (s *Server) Start(ctx context.Context) error {
	s.log.Info().
		Str("port", s.config.TURNPort).
		Str("realm", s.config.TURNRealm).
		Msg("TURN server started and listening")

	<-ctx.Done()
	return nil
}

// Stop останавливает TURN сервер
func (s *Server) Stop() error {
	s.log.Info().Msg("Stopping TURN server")
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}
	return nil
}

// GetTurnURL возвращает URL TURN сервера для использования в клиентах
func (s *Server) GetTurnURL() string {
	publicIP := s.config.TURNPublicIP
	if publicIP == "" || publicIP == "0.0.0.0" {
		publicIP = "localhost"
	}
	return fmt.Sprintf("turn:%s:%s", publicIP, s.config.TURNPort)
}

// GetTurnCredentials возвращает credentials для TURN сервера
func (s *Server) GetTurnCredentials() (username, password string) {
	return s.config.TURNUsername, s.config.TURNPassword
}

// GetRealm возвращает realm TURN сервера
func (s *Server) GetRealm() string {
	return s.config.TURNRealm
}

// GetTurnServers возвращает список TURN серверов в формате для WebRTC
func (s *Server) GetTurnServers() []string {
	turnURL := s.GetTurnURL()
	username, password := s.GetTurnCredentials()
	return []string{
		fmt.Sprintf("%s?transport=udp", turnURL),
		fmt.Sprintf("%s:%s@%s?transport=tcp", username, password, turnURL),
	}
}
