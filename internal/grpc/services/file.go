package services

import (
	"context"
	"io"
	"time"

	"github.com/backend-app/backend/internal/models"
	"github.com/backend-app/backend/internal/repository"
	"github.com/backend-app/backend/internal/storage"
	filepb "github.com/backend-app/backend/pkg/proto/file"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileService struct {
	filepb.UnimplementedFileServiceServer
	fileRepo  *repository.FileRepo
	storage   *storage.LocalStorage
	chunkSize int64
}

func NewFileService(fileRepo *repository.FileRepo, storage *storage.LocalStorage) *FileService {
	return &FileService{
		fileRepo:  fileRepo,
		storage:   storage,
		chunkSize: 64 * 1024,
	}
}

func (s *FileService) UploadFile(stream filepb.FileService_UploadFileServer) error {
	var metadata *filepb.FileMetadata
	var chunks [][]byte
	var totalSize int64

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, "failed to receive chunk")
		}

		switch data := req.Data.(type) {
		case *filepb.UploadFileRequest_Metadata:
			metadata = data.Metadata
		case *filepb.UploadFileRequest_Chunk:
			chunks = append(chunks, data.Chunk.Data)
			totalSize += int64(len(data.Chunk.Data))
		}
	}

	if metadata == nil {
		return status.Error(codes.InvalidArgument, "metadata is required")
	}

	userID, err := uuid.Parse(metadata.UserId)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if totalSize != metadata.Size {
		return status.Error(codes.InvalidArgument, "file size mismatch")
	}

	file := &models.File{
		UserID:      userID,
		Name:        metadata.Name,
		Size:        metadata.Size,
		MimeType:    metadata.MimeType,
		StorageType: models.StorageTypeLocal,
	}

	if err := file.Validate(); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.fileRepo.Create(file); err != nil {
		return status.Error(codes.Internal, "failed to create file record")
	}

	storagePath, err := s.storage.SaveFile(userID, file.ID, metadata.Name, chunks)
	if err != nil {
		s.fileRepo.Delete(file.ID)
		return status.Error(codes.Internal, "failed to save file: "+err.Error())
	}

	file.StoragePath = storagePath
	if err := s.fileRepo.Update(file); err != nil {
		s.storage.DeleteFile(storagePath)
		s.fileRepo.Delete(file.ID)
		return status.Error(codes.Internal, "failed to update file record")
	}

	return stream.SendAndClose(&filepb.UploadFileResponse{
		FileId:       file.ID.String(),
		StoragePath:  file.StoragePath,
		UploadedSize: totalSize,
	})
}

func (s *FileService) DownloadFile(req *filepb.DownloadFileRequest, stream filepb.FileService_DownloadFileServer) error {
	fileID, err := uuid.Parse(req.FileId)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid file_id")
	}

	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return status.Error(codes.Internal, "failed to get file")
	}
	if file == nil {
		return status.Error(codes.NotFound, "file not found")
	}

	userID, err := uuid.Parse(req.UserId)
	if err == nil && file.UserID != userID {
		return status.Error(codes.PermissionDenied, "file belongs to another user")
	}

	if file.StoragePath == "" {
		return status.Error(codes.NotFound, "file storage path not found")
	}

	reader, fileSize, err := s.storage.ReadFile(file.StoragePath, req.Offset, req.Limit)
	if err != nil {
		return status.Error(codes.Internal, "failed to read file: "+err.Error())
	}
	defer reader.Close()

	chunkSize := s.chunkSize
	if req.Limit > 0 && req.Limit < chunkSize {
		chunkSize = req.Limit
	}

	buffer := make([]byte, chunkSize)
	totalSent := int64(0)
	limit := req.Limit

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, "failed to read chunk: "+err.Error())
		}

		if limit > 0 && totalSent >= limit {
			break
		}

		chunkToSend := buffer[:n]
		if limit > 0 && totalSent+int64(n) > limit {
			chunkToSend = buffer[:limit-totalSent]
		}

		isLast := totalSent+int64(len(chunkToSend)) >= fileSize || (limit > 0 && totalSent+int64(len(chunkToSend)) >= limit)

		if err := stream.Send(&filepb.DownloadFileResponse{
			Data:      chunkToSend,
			ChunkSize: int64(len(chunkToSend)),
			IsLast:    isLast,
		}); err != nil {
			return status.Error(codes.Internal, "failed to send chunk: "+err.Error())
		}

		totalSent += int64(len(chunkToSend))
		if isLast {
			break
		}
	}

	return nil
}

func (s *FileService) GetFileMetadata(ctx context.Context, req *filepb.GetFileMetadataRequest) (*filepb.GetFileMetadataResponse, error) {
	fileID, err := uuid.Parse(req.FileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get file")
	}
	if file == nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	userID, err := uuid.Parse(req.UserId)
	if err == nil && file.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "file belongs to another user")
	}

	var expiresAt string
	if file.ExpiresAt != nil {
		expiresAt = file.ExpiresAt.Format(time.RFC3339)
	}

	return &filepb.GetFileMetadataResponse{
		File: &filepb.FileInfo{
			Id:          file.ID.String(),
			UserId:      file.UserID.String(),
			Name:        file.Name,
			Size:        file.Size,
			MimeType:    file.MimeType,
			StoragePath: file.StoragePath,
			StorageType: string(file.StorageType),
			ExpiresAt:   expiresAt,
			CreatedAt:   file.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   file.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *FileService) ListFiles(ctx context.Context, req *filepb.ListFilesRequest) (*filepb.ListFilesResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	files, err := s.fileRepo.GetByUserID(userID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list files")
	}

	pbFiles := make([]*filepb.FileInfo, len(files))
	for i, file := range files {
		var expiresAt string
		if file.ExpiresAt != nil {
			expiresAt = file.ExpiresAt.Format(time.RFC3339)
		}

		pbFiles[i] = &filepb.FileInfo{
			Id:          file.ID.String(),
			UserId:      file.UserID.String(),
			Name:        file.Name,
			Size:        file.Size,
			MimeType:    file.MimeType,
			StoragePath: file.StoragePath,
			StorageType: string(file.StorageType),
			ExpiresAt:   expiresAt,
			CreatedAt:   file.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   file.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &filepb.ListFilesResponse{
		Files: pbFiles,
		Total: int32(len(pbFiles)),
	}, nil
}

func (s *FileService) DeleteFile(ctx context.Context, req *filepb.DeleteFileRequest) (*filepb.DeleteFileResponse, error) {
	fileID, err := uuid.Parse(req.FileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get file")
	}
	if file == nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	if file.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "file belongs to another user")
	}

	if file.StoragePath != "" {
		if err := s.storage.DeleteFile(file.StoragePath); err != nil {
		}
	}

	if err := s.fileRepo.Delete(fileID); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete file record")
	}

	return &filepb.DeleteFileResponse{
		Success: true,
	}, nil
}
