package repository

import (
	"database/sql"
	"time"

	"github.com/backend-app/backend/internal/models"
	"github.com/google/uuid"
)

type FileRepo struct {
	db *sql.DB
}

func NewFileRepo(db *sql.DB) *FileRepo {
	return &FileRepo{db: db}
}

func (r *FileRepo) Create(file *models.File) error {
	query := `
		INSERT INTO files (id, user_id, name, size, mime_type, storage_path, storage_type, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	file.ID = uuid.New()
	now := time.Now()
	file.CreatedAt = now
	file.UpdatedAt = now

	_, err := r.db.Exec(query,
		file.ID,
		file.UserID,
		file.Name,
		file.Size,
		file.MimeType,
		file.StoragePath,
		file.StorageType,
		file.ExpiresAt,
		file.CreatedAt,
		file.UpdatedAt,
	)

	return err
}

func (r *FileRepo) GetByID(id uuid.UUID) (*models.File, error) {
	query := `
		SELECT id, user_id, name, size, mime_type, storage_path, storage_type, expires_at, created_at, updated_at
		FROM files
		WHERE id = $1
	`

	file := &models.File{}
	var expiresAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&file.ID,
		&file.UserID,
		&file.Name,
		&file.Size,
		&file.MimeType,
		&file.StoragePath,
		&file.StorageType,
		&expiresAt,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		file.ExpiresAt = &expiresAt.Time
	}

	return file, nil
}

func (r *FileRepo) GetByUserID(userID uuid.UUID, limit, offset int) ([]*models.File, error) {
	query := `
		SELECT id, user_id, name, size, mime_type, storage_path, storage_type, expires_at, created_at, updated_at
		FROM files
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var files []*models.File

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		file := &models.File{}
		var expiresAt sql.NullTime

		err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Name,
			&file.Size,
			&file.MimeType,
			&file.StoragePath,
			&file.StorageType,
			&expiresAt,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			file.ExpiresAt = &expiresAt.Time
		}

		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *FileRepo) Update(file *models.File) error {
	query := `
		UPDATE files
		SET name = $1, size = $2, mime_type = $3, storage_path = $4, storage_type = $5, expires_at = $6, updated_at = $7
		WHERE id = $8
	`

	now := time.Now()
	file.UpdatedAt = now

	res, err := r.db.Exec(query,
		file.Name,
		file.Size,
		file.MimeType,
		file.StoragePath,
		file.StorageType,
		file.ExpiresAt,
		file.UpdatedAt,
		file.ID,
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

func (r *FileRepo) Delete(id uuid.UUID) error {
	query := `
		DELETE FROM files
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

func (r *FileRepo) GetExpiredFiles() ([]*models.File, error) {
	query := `
		SELECT id, user_id, name, size, mime_type, storage_path, storage_type, expires_at, created_at, updated_at
		FROM files
		WHERE expires_at IS NOT NULL AND expires_at < NOW()
		ORDER BY expires_at ASC
	`

	var files []*models.File

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		file := &models.File{}
		var expiresAt sql.NullTime

		err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Name,
			&file.Size,
			&file.MimeType,
			&file.StoragePath,
			&file.StorageType,
			&expiresAt,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			file.ExpiresAt = &expiresAt.Time
		}

		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
