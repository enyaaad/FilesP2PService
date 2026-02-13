package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/backend-app/backend/internal/api/handlers"
	"github.com/backend-app/backend/internal/api/middleware"
	"github.com/backend-app/backend/internal/webrtc"
	"github.com/backend-app/backend/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	config      *config.Config
	db          *sql.DB
	redis       *redis.Client
	grpcClients *GRPCClients
	router      *gin.Engine
	httpServer  *http.Server
}

func NewServer(cfg *config.Config, db *sql.DB, redisClient *redis.Client, grpcClients *GRPCClients, turnServer *webrtc.Server) *Server {
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(corsMiddleware())
	router.Use(loggingMiddleware())

	router.GET("/health", healthCheck)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authHandler := handlers.NewAuthHandler(grpcClients.Auth)
	deviceHandler := handlers.NewDeviceHandler(grpcClients.Device)
	fileHandler := handlers.NewFileHandler(grpcClients.File)
	var webrtcHandler *handlers.WebRTCHandler
	if turnServer != nil {
		webrtcHandler = handlers.NewWebRTCHandler(turnServer)
	}

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(grpcClients.Auth))
		{
			devices := protected.Group("/devices")
			{
				devices.POST("", deviceHandler.Register)
				devices.GET("", deviceHandler.List)
				devices.GET("/:id", deviceHandler.Get)
				devices.PUT("/:id", deviceHandler.Update)
				devices.DELETE("/:id", deviceHandler.Delete)
				devices.POST("/:id/last-seen", deviceHandler.UpdateLastSeen)
			}

			files := protected.Group("/files")
			{
				files.POST("", fileHandler.Upload)
				files.GET("", fileHandler.List)
				files.GET("/:id", fileHandler.GetMetadata)
				files.GET("/:id/download", fileHandler.Download)
				files.DELETE("/:id", fileHandler.Delete)
			}

			if webrtcHandler != nil {
				webrtc := protected.Group("/webrtc")
				{
					webrtc.GET("/turn-credentials", webrtcHandler.GetTurnCredentials)
				}
			}
		}
	}

	server := &Server{
		config:      cfg,
		db:          db,
		redis:       redisClient,
		grpcClients: grpcClients,
		router:      router,
	}

	return server
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", s.config.Server.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("🚀 API Server starting on port %s\n", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "backend-api",
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return gin.Logger()
}
