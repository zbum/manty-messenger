package repository

import (
	"context"
	"database/sql"

	"Mmessenger/internal/models"
)

type DeviceTokenRepository struct {
	db *sql.DB
}

func NewDeviceTokenRepository(db *sql.DB) *DeviceTokenRepository {
	return &DeviceTokenRepository{db: db}
}

func (r *DeviceTokenRepository) Upsert(ctx context.Context, token *models.DeviceToken) error {
	query := `
		INSERT INTO device_tokens (user_id, device_id, platform, token)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			platform = VALUES(platform),
			token = VALUES(token),
			updated_at = NOW()
	`
	result, err := r.db.ExecContext(ctx, query, token.UserID, token.DeviceID, token.Platform, token.Token)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil && id > 0 {
		token.ID = uint64(id)
	}
	return nil
}

func (r *DeviceTokenRepository) GetByUserID(ctx context.Context, userID uint64) ([]*models.DeviceToken, error) {
	query := `
		SELECT id, user_id, device_id, platform, token, created_at, updated_at
		FROM device_tokens
		WHERE user_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*models.DeviceToken
	for rows.Next() {
		t := &models.DeviceToken{}
		err := rows.Scan(&t.ID, &t.UserID, &t.DeviceID, &t.Platform, &t.Token, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func (r *DeviceTokenRepository) GetByUserIDs(ctx context.Context, userIDs []uint64) ([]*models.DeviceToken, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, user_id, device_id, platform, token, created_at, updated_at
		FROM device_tokens
		WHERE user_id IN (?` + repeatPlaceholder(len(userIDs)-1) + `)
	`

	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*models.DeviceToken
	for rows.Next() {
		t := &models.DeviceToken{}
		err := rows.Scan(&t.ID, &t.UserID, &t.DeviceID, &t.Platform, &t.Token, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func (r *DeviceTokenRepository) DeleteByUserAndDevice(ctx context.Context, userID uint64, deviceID string) error {
	query := `DELETE FROM device_tokens WHERE user_id = ? AND device_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID, deviceID)
	return err
}

func (r *DeviceTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	query := `DELETE FROM device_tokens WHERE token = ?`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}
