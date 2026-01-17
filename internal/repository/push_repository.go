package repository

import (
	"context"
	"database/sql"

	"Mmessenger/internal/models"
)

type PushRepository struct {
	db *sql.DB
}

func NewPushRepository(db *sql.DB) *PushRepository {
	return &PushRepository{db: db}
}

// Create creates a new push subscription
func (r *PushRepository) Create(ctx context.Context, sub *models.PushSubscription) error {
	query := `
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			p256dh = VALUES(p256dh),
			auth = VALUES(auth),
			updated_at = NOW()
	`
	result, err := r.db.ExecContext(ctx, query, sub.UserID, sub.Endpoint, sub.P256dh, sub.Auth)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil && id > 0 {
		sub.ID = uint64(id)
	}
	return nil
}

// GetByUserID returns all push subscriptions for a user
func (r *PushRepository) GetByUserID(ctx context.Context, userID uint64) ([]*models.PushSubscription, error) {
	query := `
		SELECT id, user_id, endpoint, p256dh, auth, created_at, updated_at
		FROM push_subscriptions
		WHERE user_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.PushSubscription
	for rows.Next() {
		sub := &models.PushSubscription{}
		err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.P256dh, &sub.Auth, &sub.CreatedAt, &sub.UpdatedAt)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

// GetByUserIDs returns all push subscriptions for multiple users
func (r *PushRepository) GetByUserIDs(ctx context.Context, userIDs []uint64) ([]*models.PushSubscription, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, user_id, endpoint, p256dh, auth, created_at, updated_at
		FROM push_subscriptions
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

	var subs []*models.PushSubscription
	for rows.Next() {
		sub := &models.PushSubscription{}
		err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.P256dh, &sub.Auth, &sub.CreatedAt, &sub.UpdatedAt)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

// DeleteByUserAndEndpoint deletes a push subscription by user ID and endpoint
func (r *PushRepository) DeleteByUserAndEndpoint(ctx context.Context, userID uint64, endpoint string) error {
	query := `DELETE FROM push_subscriptions WHERE user_id = ? AND endpoint = ?`
	_, err := r.db.ExecContext(ctx, query, userID, endpoint)
	return err
}

// DeleteByUserID deletes all push subscriptions for a user
func (r *PushRepository) DeleteByUserID(ctx context.Context, userID uint64) error {
	query := `DELETE FROM push_subscriptions WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DeleteByEndpoint deletes a push subscription by endpoint
func (r *PushRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	query := `DELETE FROM push_subscriptions WHERE endpoint = ?`
	_, err := r.db.ExecContext(ctx, query, endpoint)
	return err
}

func repeatPlaceholder(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += ", ?"
	}
	return result
}
