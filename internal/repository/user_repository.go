package repository

import (
	"context"
	"database/sql"

	"Mmessenger/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (keycloak_id, email, username, password_hash, status)
		VALUES (?, ?, ?, ?, ?)
	`
	var keycloakID interface{}
	if user.KeycloakID.Valid {
		keycloakID = user.KeycloakID.String
	}
	result, err := r.db.ExecContext(ctx, query, keycloakID, user.Email, user.Username, user.PasswordHash, models.UserStatusOffline)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = uint64(id)
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	query := `
		SELECT id, keycloak_id, email, username, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users WHERE id = ?
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.Username, &user.PasswordHash,
		&user.AvatarURL, &user.Status, &user.LastSeenAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, keycloak_id, email, username, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users WHERE email = ?
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.Username, &user.PasswordHash,
		&user.AvatarURL, &user.Status, &user.LastSeenAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, keycloak_id, email, username, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users WHERE username = ?
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.Username, &user.PasswordHash,
		&user.AvatarURL, &user.Status, &user.LastSeenAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByKeycloakID(ctx context.Context, keycloakID string) (*models.User, error) {
	query := `
		SELECT id, keycloak_id, email, username, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users WHERE keycloak_id = ?
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, keycloakID).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.Username, &user.PasswordHash,
		&user.AvatarURL, &user.Status, &user.LastSeenAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, userID uint64, status models.UserStatus) error {
	query := `UPDATE users SET status = ?, last_seen_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, userID)
	return err
}

func (r *UserRepository) UpdateKeycloakID(ctx context.Context, userID uint64, keycloakID string) error {
	query := `UPDATE users SET keycloak_id = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, keycloakID, userID)
	return err
}

func (r *UserRepository) Search(ctx context.Context, keyword string, limit int) ([]*models.User, error) {
	query := `
		SELECT id, keycloak_id, email, username, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users
		WHERE username LIKE ? OR email LIKE ?
		LIMIT ?
	`
	searchPattern := "%" + keyword + "%"
	rows, err := r.db.QueryContext(ctx, query, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.KeycloakID, &user.Email, &user.Username, &user.PasswordHash,
			&user.AvatarURL, &user.Status, &user.LastSeenAt,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
