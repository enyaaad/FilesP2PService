package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TransferType string

const (
	TransferTypeP2P   TransferType = "p2p"
	TransferTypeCloud TransferType = "cloud"
)

type TransferStatus string

const (
	TransferStatusPending    TransferStatus = "pending"
	TransferStatusInProgress TransferStatus = "in_progress"
	TransferStatusCompleted  TransferStatus = "completed"
	TransferStatusFailed     TransferStatus = "failed"
)

type Transfer struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	FileID       uuid.UUID      `json:"file_id" db:"file_id"`
	FromDeviceID *uuid.UUID     `json:"from_device_id,omitempty" db:"from_device_id"`
	ToDeviceID   *uuid.UUID     `json:"to_device_id,omitempty" db:"to_device_id"`
	TransferType TransferType   `json:"transfer_type" db:"transfer_type"`
	Status       TransferStatus `json:"status" db:"status"`
	Progress     int64          `json:"progress" db:"progress"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

func (t *Transfer) Validate() error {
	if t.TransferType != TransferTypeP2P && t.TransferType != TransferTypeCloud {
		return errors.New("invalid transfer type")
	}
	if t.Status != TransferStatusPending &&
		t.Status != TransferStatusInProgress &&
		t.Status != TransferStatusCompleted &&
		t.Status != TransferStatusFailed {
		return errors.New("invalid transfer status")
	}
	if t.Progress < 0 {
		return errors.New("progress cannot be negative")
	}
	return nil
}

func (t *Transfer) UpdateProgress(progress int64) {
	t.Progress = progress
	t.UpdatedAt = time.Now()
	if progress >= 100 {
		t.Status = TransferStatusCompleted
	}
}
