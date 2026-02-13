package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeR2    StorageType = "r2"
)

type File struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	UserID      uuid.UUID   `json:"user_id" db:"user_id"`
	Name        string      `json:"name" db:"name"`
	Size        int64       `json:"size" db:"size"`
	MimeType    string      `json:"mime_type" db:"mime_type"`
	StoragePath string      `json:"storage_path" db:"storage_path"`
	StorageType StorageType `json:"storage_type" db:"storage_type"`
	ExpiresAt   *time.Time  `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

func (f *File) Validate() error {
	if f.Name == "" {
		return errors.New("file name is required")
	}
	if f.Size <= 0 {
		return errors.New("file size must be greater than 0")
	}
	if f.StoragePath == "" {
		return errors.New("storage path is required")
	}
	if f.StorageType != StorageTypeLocal && f.StorageType != StorageTypeR2 {
		return errors.New("invalid storage type")
	}
	return nil
}

func (f *File) IsExpired() bool {
	if f.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*f.ExpiresAt)
}
