package services

import (
	"context"
	"time"

	"github.com/backend-app/backend/internal/models"
	"github.com/backend-app/backend/internal/repository"
	transferpb "github.com/backend-app/backend/pkg/proto/transfer"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransferService struct {
	transferpb.UnimplementedTransferServiceServer
	transferRepo *repository.TransferRepo
}

func NewTransferService(transferRepo *repository.TransferRepo) *TransferService {
	return &TransferService{
		transferRepo: transferRepo,
	}
}

func (s *TransferService) CreateTransfer(ctx context.Context, req *transferpb.CreateTransferRequest) (*transferpb.CreateTransferResponse, error) {
	fileID, err := uuid.Parse(req.FileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	transferType := models.TransferType(req.TransferType)
	if transferType != models.TransferTypeP2P && transferType != models.TransferTypeCloud {
		return nil, status.Error(codes.InvalidArgument, "invalid transfer_type")
	}

	transfer := &models.Transfer{
		FileID:       fileID,
		TransferType: transferType,
		Status:       models.TransferStatusPending,
		Progress:     0,
	}

	if req.FromDeviceId != "" {
		fromDeviceID, err := uuid.Parse(req.FromDeviceId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid from_device_id")
		}
		transfer.FromDeviceID = &fromDeviceID
	}

	if req.ToDeviceId != "" {
		toDeviceID, err := uuid.Parse(req.ToDeviceId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid to_device_id")
		}
		transfer.ToDeviceID = &toDeviceID
	}

	if err := transfer.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.transferRepo.Create(transfer); err != nil {
		return nil, status.Error(codes.Internal, "failed to create transfer")
	}

	return &transferpb.CreateTransferResponse{
		Transfer: s.transferToProto(transfer),
	}, nil
}

func (s *TransferService) GetTransfer(ctx context.Context, req *transferpb.GetTransferRequest) (*transferpb.GetTransferResponse, error) {
	transferID, err := uuid.Parse(req.TransferId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid transfer_id")
	}

	transfer, err := s.transferRepo.GetByID(transferID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get transfer")
	}
	if transfer == nil {
		return nil, status.Error(codes.NotFound, "transfer not found")
	}

	return &transferpb.GetTransferResponse{
		Transfer: s.transferToProto(transfer),
	}, nil
}

func (s *TransferService) UpdateTransferStatus(ctx context.Context, req *transferpb.UpdateTransferStatusRequest) (*transferpb.UpdateTransferStatusResponse, error) {
	transferID, err := uuid.Parse(req.TransferId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid transfer_id")
	}

	transferStatus := models.TransferStatus(req.Status)
	if transferStatus != models.TransferStatusPending &&
		transferStatus != models.TransferStatusInProgress &&
		transferStatus != models.TransferStatusCompleted &&
		transferStatus != models.TransferStatusFailed {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	if err := s.transferRepo.UpdateStatus(transferID, transferStatus, req.Progress); err != nil {
		return nil, status.Error(codes.Internal, "failed to update transfer status")
	}

	transfer, err := s.transferRepo.GetByID(transferID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get transfer")
	}

	return &transferpb.UpdateTransferStatusResponse{
		Transfer: s.transferToProto(transfer),
	}, nil
}

func (s *TransferService) ListTransfers(ctx context.Context, req *transferpb.ListTransfersRequest) (*transferpb.ListTransfersResponse, error) {
	var transfers []*models.Transfer
	var err error

	if req.FileId != "" {
		fileID, err := uuid.Parse(req.FileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid file_id")
		}
		transfers, err = s.transferRepo.GetByFileID(fileID)
	} else if req.Status != "" {
		transferStatus := models.TransferStatus(req.Status)
		transfers, err = s.transferRepo.GetByStatus(transferStatus)
	} else {
		return nil, status.Error(codes.InvalidArgument, "file_id or status must be provided")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list transfers")
	}

	pbTransfers := make([]*transferpb.Transfer, len(transfers))
	for i, transfer := range transfers {
		pbTransfers[i] = s.transferToProto(transfer)
	}

	return &transferpb.ListTransfersResponse{
		Transfers: pbTransfers,
	}, nil
}

func (s *TransferService) StreamTransferProgress(req *transferpb.StreamTransferProgressRequest, stream transferpb.TransferService_StreamTransferProgressServer) error {
	transferID, err := uuid.Parse(req.TransferId)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid transfer_id")
	}

	transfer, err := s.transferRepo.GetByID(transferID)
	if err != nil {
		return status.Error(codes.Internal, "failed to get transfer")
	}
	if transfer == nil {
		return status.Error(codes.NotFound, "transfer not found")
	}

	return stream.Send(&transferpb.TransferProgressUpdate{
		TransferId: transfer.ID.String(),
		Status:     string(transfer.Status),
		Progress:   transfer.Progress,
		TotalSize:  0, // TODO: получить из файла
	})
}

func (s *TransferService) transferToProto(transfer *models.Transfer) *transferpb.Transfer {
	pbTransfer := &transferpb.Transfer{
		Id:           transfer.ID.String(),
		FileId:       transfer.FileID.String(),
		TransferType: string(transfer.TransferType),
		Status:       string(transfer.Status),
		Progress:     transfer.Progress,
		CreatedAt:    transfer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    transfer.UpdatedAt.Format(time.RFC3339),
	}

	if transfer.FromDeviceID != nil {
		pbTransfer.FromDeviceId = transfer.FromDeviceID.String()
	}

	if transfer.ToDeviceID != nil {
		pbTransfer.ToDeviceId = transfer.ToDeviceID.String()
	}

	return pbTransfer
}
