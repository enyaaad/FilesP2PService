package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/backend-app/backend/internal/api"
	"github.com/backend-app/backend/internal/database"
	"github.com/backend-app/backend/internal/grpc"
	"github.com/backend-app/backend/internal/repository"
	"github.com/backend-app/backend/internal/webrtc"
	"github.com/backend-app/backend/internal/websocket"
	"github.com/backend-app/backend/pkg/config"
	"github.com/backend-app/backend/pkg/logger"

	_ "github.com/backend-app/backend/docs"
)

// @title Backend API
// @version 1.0
// @description Backend для файлообмена между устройствами с поддержкой P2P через WebRTC
//
// ## Основные возможности:
// - Регистрация и аутентификация пользователей
// - Управление устройствами
// - Загрузка и скачивание файлов с поддержкой Range requests
// - WebRTC Signaling для P2P передачи файлов
// - TURN сервер для ретрансляции трафика
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	logger.Init(cfg.Server.Environment)
	log := logger.Get()

	log.Info().Msg("Starting backend server...")

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("Connected to PostgreSQL")

	if err := database.RunMigrations(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	log.Info().Msg("Database migrations completed")

	redisClient, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	log.Info().Msg("Connected to Redis")

	grpcServer := grpc.NewServer(cfg, db, redisClient)
	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start gRPC server")
		}
	}()
	log.Info().Str("port", cfg.Server.GRPCPort).Msg("gRPC server started")

	grpcAddr := fmt.Sprintf("localhost:%s", cfg.Server.GRPCPort)
	grpcClients, err := api.NewGRPCClients(grpcAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create gRPC clients")
	}
	defer grpcClients.Close()
	log.Info().Msg("gRPC clients connected")

	deviceRepo := repository.NewDeviceRepo(db)

	turnServer, err := webrtc.NewTurnServer(&cfg.WebRTC)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create TURN server")
	}
	defer turnServer.Stop()

	turnCtx, turnCancel := context.WithCancel(context.Background())
	defer turnCancel()
	go func() {
		if err := turnServer.Start(turnCtx); err != nil {
			log.Error().Err(err).Msg("TURN server error")
		}
	}()
	log.Info().
		Str("port", cfg.WebRTC.TURNPort).
		Str("url", turnServer.GetTurnURL()).
		Msg("TURN server started")

	httpServer := api.NewServer(cfg, db, redisClient, grpcClients, turnServer)
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()
	log.Info().Str("port", cfg.Server.Port).Msg("HTTP server started")

	wsServer := websocket.NewServer(cfg, db, deviceRepo)
	go func() {
		if err := wsServer.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start WebSocket server")
		}
	}()
	log.Info().Str("port", cfg.Server.WebSocketPort).Msg("WebSocket signaling server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP server forced to shutdown")
	}

	grpcServer.Stop()

	if err := wsServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("WebSocket server forced to shutdown")
	}

	turnCancel()
	if err := turnServer.Stop(); err != nil {
		log.Error().Err(err).Msg("TURN server forced to shutdown")
	}

	log.Info().Msg("Servers exited")
}
