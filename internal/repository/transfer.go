package repository

import (
	"database/sql"
	"time"

	"github.com/backend-app/backend/internal/models"
	"github.com/google/uuid"
)

type TransferRepo struct {
	db *sql.DB
}

func NewTransferRepo(db *sql.DB) *TransferRepo {
	return &TransferRepo{db: db}
}

func (r *TransferRepo) Create(transfer *models.Transfer) error {
	query := `
		INSERT INTO transfers (id, file_id, from_device_id, to_device_id, transfer_type, status, progress, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	transfer.ID = uuid.New()
	now := time.Now()
	transfer.CreatedAt = now
	transfer.UpdatedAt = now

	_, err := r.db.Exec(query,
		transfer.ID,
		transfer.FileID,
		transfer.FromDeviceID,
		transfer.ToDeviceID,
		transfer.TransferType,
		transfer.Status,
		transfer.Progress,
		transfer.CreatedAt,
		transfer.UpdatedAt,
	)

	return err
}

func (r *TransferRepo) GetByID(id uuid.UUID) (*models.Transfer, error) {
	query := `
		SELECT id, file_id, from_device_id, to_device_id, transfer_type, status, progress, created_at, updated_at
		FROM transfers
		WHERE id = $1
	`

	transfer := &models.Transfer{}
	var fromDeviceID, toDeviceID sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&transfer.ID,
		&transfer.FileID,
		&fromDeviceID,
		&toDeviceID,
		&transfer.TransferType,
		&transfer.Status,
		&transfer.Progress,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if fromDeviceID.Valid {
		parsedUUID, err := uuid.Parse(fromDeviceID.String)
		if err == nil {
			transfer.FromDeviceID = &parsedUUID
		}
	}

	if toDeviceID.Valid {
		parsedUUID, err := uuid.Parse(toDeviceID.String)
		if err == nil {
			transfer.ToDeviceID = &parsedUUID
		}
	}

	return transfer, nil
}

func (r *TransferRepo) GetByFileID(fileID uuid.UUID) ([]*models.Transfer, error) {
	query := `
		SELECT id, file_id, from_device_id, to_device_id, transfer_type, status, progress, created_at, updated_at
		FROM transfers
		WHERE file_id = $1
		ORDER BY created_at DESC
	`

	var transfers []*models.Transfer

	rows, err := r.db.Query(query, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		transfer := &models.Transfer{}
		var fromDeviceID, toDeviceID sql.NullString

		err := rows.Scan(
			&transfer.ID,
			&transfer.FileID,
			&fromDeviceID,
			&toDeviceID,
			&transfer.TransferType,
			&transfer.Status,
			&transfer.Progress,
			&transfer.CreatedAt,
			&transfer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if fromDeviceID.Valid {
			parsedUUID, err := uuid.Parse(fromDeviceID.String)
			if err == nil {
				transfer.FromDeviceID = &parsedUUID
			}
		}

		if toDeviceID.Valid {
			parsedUUID, err := uuid.Parse(toDeviceID.String)
			if err == nil {
				transfer.ToDeviceID = &parsedUUID
			}
		}

		transfers = append(transfers, transfer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transfers, nil
}

func (r *TransferRepo) UpdateStatus(id uuid.UUID, status models.TransferStatus, progress int64) error {
	query := `
		UPDATE transfers
		SET status = $1, progress = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()

	res, err := r.db.Exec(query,
		status,
		progress,
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

func (r *TransferRepo) GetByStatus(status models.TransferStatus) ([]*models.Transfer, error) {
	query := `
		SELECT id, file_id, from_device_id, to_device_id, transfer_type, status, progress, created_at, updated_at
		FROM transfers
		WHERE status = $1
		ORDER BY created_at DESC
	`

	var transfers []*models.Transfer

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		transfer := &models.Transfer{}
		var fromDeviceID, toDeviceID sql.NullString

		err := rows.Scan(
			&transfer.ID,
			&transfer.FileID,
			&fromDeviceID,
			&toDeviceID,
			&transfer.TransferType,
			&transfer.Status,
			&transfer.Progress,
			&transfer.CreatedAt,
			&transfer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if fromDeviceID.Valid {
			parsedUUID, err := uuid.Parse(fromDeviceID.String)
			if err == nil {
				transfer.FromDeviceID = &parsedUUID
			}
		}

		if toDeviceID.Valid {
			parsedUUID, err := uuid.Parse(toDeviceID.String)
			if err == nil {
				transfer.ToDeviceID = &parsedUUID
			}
		}

		transfers = append(transfers, transfer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transfers, nil
}

func (r *TransferRepo) Update(transfer *models.Transfer) error {
	query := `
		UPDATE transfers
		SET file_id = $1, from_device_id = $2, to_device_id = $3, transfer_type = $4, status = $5, progress = $6, updated_at = $7
		WHERE id = $8
	`

	now := time.Now()
	transfer.UpdatedAt = now

	res, err := r.db.Exec(query,
		transfer.FileID,
		transfer.FromDeviceID,
		transfer.ToDeviceID,
		transfer.TransferType,
		transfer.Status,
		transfer.Progress,
		transfer.UpdatedAt,
		transfer.ID,
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

func (r *TransferRepo) Delete(id uuid.UUID) error {
	query := `
		DELETE FROM transfers
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
