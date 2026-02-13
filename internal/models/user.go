package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

func (u *User) Scan(value interface{}) error {
	if value == nil {
		return errors.New("user scan: value is nil")
	}

	if id, ok := value.([]byte); ok {
		return u.ID.UnmarshalBinary(id)
	}

	return errors.New("user scan: invalid type")
}

func (u *User) Value() (driver.Value, error) {
	return u.ID.String(), nil
}

func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	return nil
}
