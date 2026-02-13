package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// SaveFile сохраняет файл в хранилище
// Возвращает относительный путь к файлу (относительно basePath)
func (s *LocalStorage) SaveFile(userID uuid.UUID, fileID uuid.UUID, filename string, chunks [][]byte) (string, error) {
	userDir := filepath.Join(s.basePath, "users", userID.String(), "files")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	filePath := filepath.Join(userDir, fmt.Sprintf("%s_%s", fileID.String(), filename))

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for _, chunk := range chunks {
		if _, err := file.Write(chunk); err != nil {
			return "", fmt.Errorf("failed to write chunk: %w", err)
		}
	}

	relPath, err := filepath.Rel(s.basePath, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	return relPath, nil
}

// ReadFile читает файл из хранилища
// offset и limit для поддержки Range requests
func (s *LocalStorage) ReadFile(storagePath string, offset, limit int64) (io.ReadCloser, int64, error) {
	fullPath := filepath.Join(s.basePath, storagePath)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, fmt.Errorf("file not found")
		}
		return nil, 0, fmt.Errorf("failed to open file: %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := fileInfo.Size()

	if offset > 0 {
		if _, err := file.Seek(offset, 0); err != nil {
			file.Close()
			return nil, 0, fmt.Errorf("failed to seek file: %w", err)
		}
	}

	var reader io.Reader = file
	if limit > 0 && offset+limit < fileSize {
		reader = io.LimitReader(file, limit)
	}

	return &readCloserWrapper{
		Reader: reader,
		closer: file,
	}, fileSize, nil
}

// DeleteFile удаляет файл из хранилища
func (s *LocalStorage) DeleteFile(storagePath string) error {
	fullPath := filepath.Join(s.basePath, storagePath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// FileExists проверяет существование файла
func (s *LocalStorage) FileExists(storagePath string) bool {
	fullPath := filepath.Join(s.basePath, storagePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// readCloserWrapper обертка для io.Reader, чтобы сделать его io.ReadCloser
type readCloserWrapper struct {
	io.Reader
	closer io.Closer
}

func (r *readCloserWrapper) Close() error {
	return r.closer.Close()
}
