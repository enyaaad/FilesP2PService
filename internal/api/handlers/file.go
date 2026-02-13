package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/backend-app/backend/internal/api/middleware"
	filepb "github.com/backend-app/backend/pkg/proto/file"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileHandler struct {
	fileClient filepb.FileServiceClient
}

func NewFileHandler(fileClient filepb.FileServiceClient) *FileHandler {
	return &FileHandler{
		fileClient: fileClient,
	}
}

type FileResponse struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID      string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string `json:"name" example:"document.pdf"`
	Size        int64  `json:"size" example:"1024000"`
	MimeType    string `json:"mime_type" example:"application/pdf"`
	StoragePath string `json:"storage_path" example:"users/550e8400-e29b-41d4-a716-446655440000/files/..."`
	StorageType string `json:"storage_type" example:"local"`
	ExpiresAt   string `json:"expires_at,omitempty" example:"2024-02-01T00:00:00Z"`
	CreatedAt   string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

type ListFilesResponse struct {
	Files []FileResponse `json:"files"`
	Total int32          `json:"total" example:"10"`
}

type UploadFileResponse struct {
	FileID       string `json:"file_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	StoragePath  string `json:"storage_path" example:"users/550e8400-e29b-41d4-a716-446655440000/files/..."`
	UploadedSize int64  `json:"uploaded_size" example:"1024000"`
}

// Upload godoc
// @Summary Загрузка файла
// @Description Загружает файл на сервер через multipart/form-data. Поддерживает потоковую загрузку больших файлов.
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Файл для загрузки"
// @Success 201 {object} UploadFileResponse "Файл успешно загружен"
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /files [post]
func (h *FileHandler) Upload(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse multipart form"})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	fileHeader := files[0]
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	stream, err := h.fileClient.UploadFile(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload stream"})
		return
	}

	fileSize := fileHeader.Size
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	err = stream.Send(&filepb.UploadFileRequest{
		Data: &filepb.UploadFileRequest_Metadata{
			Metadata: &filepb.FileMetadata{
				Name:     fileHeader.Filename,
				Size:     fileSize,
				MimeType: mimeType,
				UserId:   userID.String(),
			},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send metadata"})
		return
	}

	buffer := make([]byte, 64*1024)
	chunkNumber := 0
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file chunk"})
			return
		}

		err = stream.Send(&filepb.UploadFileRequest{
			Data: &filepb.UploadFileRequest_Chunk{
				Chunk: &filepb.FileChunk{
					Data:        buffer[:n],
					ChunkNumber: int32(chunkNumber),
				},
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send chunk"})
			return
		}

		chunkNumber++
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	c.JSON(http.StatusCreated, UploadFileResponse{
		FileID:       resp.FileId,
		StoragePath:  resp.StoragePath,
		UploadedSize: resp.UploadedSize,
	})
}

// Download godoc
// @Summary Скачивание файла
// @Description Скачивает файл с сервера (поддерживает Range requests для частичного скачивания). Использует потоковую передачу для больших файлов.
// @Tags files
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param id path string true "ID файла" format(uuid)
// @Param offset query int false "Смещение в байтах для частичного скачивания" default(0) example:"0"
// @Param limit query int false "Лимит байт для чтения (0 = весь файл)" default(0) example:"1048576"
// @Success 200 {file} file "Файл (бинарные данные)"
// @Failure 400 {object} map[string]string "Неверный ID файла" example:"{\"error\":\"file_id is required\"}"
// @Failure 401 {object} map[string]string "Не авторизован" example:"{\"error\":\"unauthorized\"}"
// @Failure 403 {object} map[string]string "Нет доступа к файлу" example:"{\"error\":\"permission denied\"}"
// @Failure 404 {object} map[string]string "Файл не найден" example:"{\"error\":\"file not found\"}"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера" example:"{\"error\":\"failed to download file\"}"
// @Router /files/{id}/download [get]
func (h *FileHandler) Download(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id is required"})
		return
	}

	var offset, limit int64
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, _ = strconv.ParseInt(offsetStr, 10, 64)
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, _ = strconv.ParseInt(limitStr, 10, 64)
	}

	stream, err := h.fileClient.DownloadFile(context.Background(), &filepb.DownloadFileRequest{
		FileId: fileID,
		UserId: userID.String(),
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			case codes.PermissionDenied:
				c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download file"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download file"})
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment")
	c.Status(http.StatusOK)

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to receive chunk"})
			return
		}

		if _, err := c.Writer.Write(resp.Data); err != nil {
			return
		}
		c.Writer.Flush()

		if resp.IsLast {
			break
		}
	}
}

// GetMetadata godoc
// @Summary Метаданные файла
// @Description Возвращает метаданные файла
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID файла" format(uuid)
// @Success 200 {object} FileResponse "Метаданные файла"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Нет доступа к файлу"
// @Failure 404 {object} map[string]string "Файл не найден"
// @Router /files/{id} [get]
func (h *FileHandler) GetMetadata(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id is required"})
		return
	}

	resp, err := h.fileClient.GetFileMetadata(context.Background(), &filepb.GetFileMetadataRequest{
		FileId: fileID,
		UserId: userID.String(),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			case codes.PermissionDenied:
				c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file metadata"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file metadata"})
		return
	}

	c.JSON(http.StatusOK, FileResponse{
		ID:          resp.File.Id,
		UserID:      resp.File.UserId,
		Name:        resp.File.Name,
		Size:        resp.File.Size,
		MimeType:    resp.File.MimeType,
		StoragePath: resp.File.StoragePath,
		StorageType: resp.File.StorageType,
		ExpiresAt:   resp.File.ExpiresAt,
		CreatedAt:   resp.File.CreatedAt,
		UpdatedAt:   resp.File.UpdatedAt,
	})
}

// List godoc
// @Summary Список файлов
// @Description Возвращает список файлов пользователя с пагинацией
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Лимит файлов" default(50)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} ListFilesResponse "Список файлов"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Router /files [get]
func (h *FileHandler) List(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := int32(50)
	offset := int32(0)

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(l)
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
			offset = int32(o)
		}
	}

	resp, err := h.fileClient.ListFiles(context.Background(), &filepb.ListFilesRequest{
		UserId: userID.String(),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list files"})
		return
	}

	files := make([]FileResponse, len(resp.Files))
	for i, file := range resp.Files {
		files[i] = FileResponse{
			ID:          file.Id,
			UserID:      file.UserId,
			Name:        file.Name,
			Size:        file.Size,
			MimeType:    file.MimeType,
			StoragePath: file.StoragePath,
			StorageType: file.StorageType,
			ExpiresAt:   file.ExpiresAt,
			CreatedAt:   file.CreatedAt,
			UpdatedAt:   file.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, ListFilesResponse{
		Files: files,
		Total: resp.Total,
	})
}

// Delete godoc
// @Summary Удаление файла
// @Description Удаляет файл с сервера
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID файла" format(uuid)
// @Success 200 {object} map[string]bool "Файл удален"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Нет доступа к файлу"
// @Failure 404 {object} map[string]string "Файл не найден"
// @Router /files/{id} [delete]
func (h *FileHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id is required"})
		return
	}

	resp, err := h.fileClient.DeleteFile(context.Background(), &filepb.DeleteFileRequest{
		FileId: fileID,
		UserId: userID.String(),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			case codes.PermissionDenied:
				c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
	})
}
