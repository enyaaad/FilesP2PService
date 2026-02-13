package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type DeviceType string

const (
	DeviceTypeDesktop DeviceType = "desktop"
	DeviceTypeMobile  DeviceType = "mobile"
)

type Device struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	Name        string     `json:"name" db:"name"`
	DeviceType  DeviceType `json:"device_type" db:"device_type"`
	DeviceToken string     `json:"device_token" db:"device_token"`
	LastSeenAt  time.Time  `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

func (d *Device) Validate() error {
	if d.Name == "" {
		return errors.New("device name is required")
	}
	if d.DeviceType != DeviceTypeDesktop && d.DeviceType != DeviceTypeMobile {
		return errors.New("invalid device type")
	}
	if d.DeviceToken == "" {
		return errors.New("device token is required")
	}
	return nil
}
