package repository

import (
	"database/sql"
	"time"

	"github.com/backend-app/backend/internal/models"
	"github.com/google/uuid"
)

type DeviceRepo struct {
	db *sql.DB
}

func NewDeviceRepo(db *sql.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) Create(device *models.Device) error {
	query := `
		INSERT INTO devices (id, user_id, name, device_type, device_token, last_seen_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	device.ID = uuid.New()
	now := time.Now()
	device.CreatedAt = now
	device.UpdatedAt = now
	device.LastSeenAt = now

	_, err := r.db.Exec(query,
		device.ID,
		device.UserID,
		device.Name,
		device.DeviceType,
		device.DeviceToken,
		device.LastSeenAt,
		device.CreatedAt,
		device.UpdatedAt,
	)

	return err
}

func (r *DeviceRepo) GetByID(id uuid.UUID) (*models.Device, error) {
	query := `
		SELECT id, user_id, name, device_type, device_token, last_seen_at, created_at, updated_at
		FROM devices
		WHERE id = $1
	`

	device := &models.Device{}
	err := r.db.QueryRow(query, id).Scan(
		&device.ID,
		&device.UserID,
		&device.Name,
		&device.DeviceType,
		&device.DeviceToken,
		&device.LastSeenAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return device, nil
}

func (r *DeviceRepo) GetByToken(token string) (*models.Device, error) {
	query := `
		SELECT id, user_id, name, device_type, device_token, last_seen_at, created_at, updated_at
		FROM devices
		WHERE device_token = $1
	`

	device := &models.Device{}
	err := r.db.QueryRow(query, token).Scan(
		&device.ID,
		&device.UserID,
		&device.Name,
		&device.DeviceType,
		&device.DeviceToken,
		&device.LastSeenAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return device, nil
}

func (r *DeviceRepo) GetByUserID(userID uuid.UUID) ([]*models.Device, error) {
	query := `
		SELECT id, user_id, name, device_type, device_token, last_seen_at, created_at, updated_at
		FROM devices
		WHERE user_id = $1
	`

	var devices []*models.Device

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		device := &models.Device{}

		err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.Name,
			&device.DeviceType,
			&device.DeviceToken,
			&device.LastSeenAt,
			&device.CreatedAt,
			&device.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return devices, nil
}

func (r *DeviceRepo) Update(device *models.Device) error {
	query := `
		UPDATE devices
		SET name = $1, device_type = $2, device_token = $3, last_seen_at = $4, updated_at = $5
		WHERE id = $6
	`

	now := time.Now()
	device.UpdatedAt = now

	res, err := r.db.Exec(query,
		device.Name,
		device.DeviceType,
		device.DeviceToken,
		device.LastSeenAt,
		device.UpdatedAt,
		device.ID,
	)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *DeviceRepo) UpdateLastSeen(id uuid.UUID) error {
	query := `
		UPDATE devices
		SET last_seen_at = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()

	res, err := r.db.Exec(query,
		now,
		now,
		id,
	)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *DeviceRepo) Delete(id uuid.UUID) error {
	query := `
		DELETE FROM devices
		WHERE id = $1
	`

	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
