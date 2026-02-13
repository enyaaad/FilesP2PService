package api

import (
	"fmt"

	authpb "github.com/backend-app/backend/pkg/proto/auth"
	devicepb "github.com/backend-app/backend/pkg/proto/device"
	filepb "github.com/backend-app/backend/pkg/proto/file"
	transferpb "github.com/backend-app/backend/pkg/proto/transfer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClients struct {
	Auth     authpb.AuthServiceClient
	Device   devicepb.DeviceServiceClient
	File     filepb.FileServiceClient
	Transfer transferpb.TransferServiceClient
	conn     *grpc.ClientConn
}

func NewGRPCClients(grpcAddr string) (*GRPCClients, error) {
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &GRPCClients{
		Auth:     authpb.NewAuthServiceClient(conn),
		Device:   devicepb.NewDeviceServiceClient(conn),
		File:     filepb.NewFileServiceClient(conn),
		Transfer: transferpb.NewTransferServiceClient(conn),
		conn:     conn,
	}, nil
}

func (c *GRPCClients) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
