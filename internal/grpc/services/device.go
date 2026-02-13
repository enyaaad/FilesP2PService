package services

import (
	"context"
	"time"

	"github.com/backend-app/backend/internal/models"
	"github.com/backend-app/backend/internal/repository"
	devicepb "github.com/backend-app/backend/pkg/proto/device"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeviceService struct {
	devicepb.UnimplementedDeviceServiceServer
	deviceRepo *repository.DeviceRepo
}

func NewDeviceService(deviceRepo *repository.DeviceRepo) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
	}
}

func (s *DeviceService) RegisterDevice(ctx context.Context, req *devicepb.RegisterDeviceRequest) (*devicepb.RegisterDeviceResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	deviceType := models.DeviceType(req.DeviceType)
	if deviceType != models.DeviceTypeDesktop && deviceType != models.DeviceTypeMobile {
		return nil, status.Error(codes.InvalidArgument, "invalid device_type")
	}

	device := &models.Device{
		UserID:      userID,
		Name:        req.Name,
		DeviceType:  deviceType,
		DeviceToken: uuid.New().String(),
	}

	if err := device.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.deviceRepo.Create(device); err != nil {
		return nil, status.Error(codes.Internal, "failed to create device")
	}

	return &devicepb.RegisterDeviceResponse{
		Device: &devicepb.Device{
			Id:          device.ID.String(),
			UserId:      device.UserID.String(),
			Name:        device.Name,
			DeviceType:  string(device.DeviceType),
			DeviceToken: device.DeviceToken,
			LastSeenAt:  device.LastSeenAt.Format(time.RFC3339),
			CreatedAt:   device.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   device.UpdatedAt.Format(time.RFC3339),
		},
		DeviceToken: device.DeviceToken,
	}, nil
}

func (s *DeviceService) GetDevice(ctx context.Context, req *devicepb.GetDeviceRequest) (*devicepb.GetDeviceResponse, error) {
	deviceID, err := uuid.Parse(req.DeviceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid device_id")
	}

	device, err := s.deviceRepo.GetByID(deviceID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get device")
	}
	if device == nil {
		return nil, status.Error(codes.NotFound, "device not found")
	}

	userID, err := uuid.Parse(req.UserId)
	if err == nil && device.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "device belongs to another user")
	}

	return &devicepb.GetDeviceResponse{
		Device: &devicepb.Device{
			Id:          device.ID.String(),
			UserId:      device.UserID.String(),
			Name:        device.Name,
			DeviceType:  string(device.DeviceType),
			DeviceToken: device.DeviceToken,
			LastSeenAt:  device.LastSeenAt.Format(time.RFC3339),
			CreatedAt:   device.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   device.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *DeviceService) ListDevices(ctx context.Context, req *devicepb.ListDevicesRequest) (*devicepb.ListDevicesResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	devices, err := s.deviceRepo.GetByUserID(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list devices")
	}

	pbDevices := make([]*devicepb.Device, len(devices))
	for i, device := range devices {
		pbDevices[i] = &devicepb.Device{
			Id:          device.ID.String(),
			UserId:      device.UserID.String(),
			Name:        device.Name,
			DeviceType:  string(device.DeviceType),
			DeviceToken: device.DeviceToken,
			LastSeenAt:  device.LastSeenAt.Format(time.RFC3339),
			CreatedAt:   device.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   device.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &devicepb.ListDevicesResponse{
		Devices: pbDevices,
	}, nil
}

func (s *DeviceService) UpdateDevice(ctx context.Context, req *devicepb.UpdateDeviceRequest) (*devicepb.UpdateDeviceResponse, error) {
	deviceID, err := uuid.Parse(req.DeviceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid device_id")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	device, err := s.deviceRepo.GetByID(deviceID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get device")
	}
	if device == nil {
		return nil, status.Error(codes.NotFound, "device not found")
	}

	if device.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "device belongs to another user")
	}

	device.Name = req.Name
	if req.DeviceType != "" {
		device.DeviceType = models.DeviceType(req.DeviceType)
	}

	if err := device.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.deviceRepo.Update(device); err != nil {
		return nil, status.Error(codes.Internal, "failed to update device")
	}

	return &devicepb.UpdateDeviceResponse{
		Device: &devicepb.Device{
			Id:          device.ID.String(),
			UserId:      device.UserID.String(),
			Name:        device.Name,
			DeviceType:  string(device.DeviceType),
			DeviceToken: device.DeviceToken,
			LastSeenAt:  device.LastSeenAt.Format(time.RFC3339),
			CreatedAt:   device.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   device.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, req *devicepb.DeleteDeviceRequest) (*devicepb.DeleteDeviceResponse, error) {
	deviceID, err := uuid.Parse(req.DeviceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid device_id")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	device, err := s.deviceRepo.GetByID(deviceID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get device")
	}
	if device == nil {
		return nil, status.Error(codes.NotFound, "device not found")
	}

	if device.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "device belongs to another user")
	}

	if err := s.deviceRepo.Delete(deviceID); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete device")
	}

	return &devicepb.DeleteDeviceResponse{
		Success: true,
	}, nil
}

func (s *DeviceService) UpdateLastSeen(ctx context.Context, req *devicepb.UpdateLastSeenRequest) (*devicepb.UpdateLastSeenResponse, error) {
	deviceID, err := uuid.Parse(req.DeviceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid device_id")
	}

	if err := s.deviceRepo.UpdateLastSeen(deviceID); err != nil {
		return nil, status.Error(codes.Internal, "failed to update last seen")
	}

	return &devicepb.UpdateLastSeenResponse{
		Success: true,
	}, nil
}
