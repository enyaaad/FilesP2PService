package grpc

import (
	"database/sql"
	"fmt"
	"net"

	"github.com/backend-app/backend/internal/grpc/services"
	"github.com/backend-app/backend/internal/repository"
	"github.com/backend-app/backend/internal/storage"
	"github.com/backend-app/backend/pkg/config"
	authpb "github.com/backend-app/backend/pkg/proto/auth"
	devicepb "github.com/backend-app/backend/pkg/proto/device"
	filepb "github.com/backend-app/backend/pkg/proto/file"
	transferpb "github.com/backend-app/backend/pkg/proto/transfer"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	config     *config.Config
}

func NewServer(cfg *config.Config, db *sql.DB, redisClient *redis.Client) *Server {
	grpcServer := grpc.NewServer()

	userRepo := repository.NewUserRepo(db)
	deviceRepo := repository.NewDeviceRepo(db)
	fileRepo := repository.NewFileRepo(db)
	transferRepo := repository.NewTransferRepo(db)

	localStorage, err := storage.NewLocalStorage(cfg.Storage.LocalPath)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize storage: %v", err))
	}

	authpb.RegisterAuthServiceServer(grpcServer, services.NewAuthService(userRepo, cfg.Server.JWTSecret))
	devicepb.RegisterDeviceServiceServer(grpcServer, services.NewDeviceService(deviceRepo))
	filepb.RegisterFileServiceServer(grpcServer, services.NewFileService(fileRepo, localStorage))
	transferpb.RegisterTransferServiceServer(grpcServer, services.NewTransferService(transferRepo))

	return &Server{
		grpcServer: grpcServer,
		config:     cfg,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.config.Server.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	fmt.Printf("🚀 gRPC Server starting on port %s\n", s.config.Server.GRPCPort)
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
