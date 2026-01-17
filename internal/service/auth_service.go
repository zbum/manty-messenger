package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"Mmessenger/internal/models"
	"Mmessenger/internal/repository"
	"Mmessenger/pkg/keycloak"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type AuthService struct {
	userRepo *repository.UserRepository
	db       *sql.DB
}

func NewAuthService(userRepo *repository.UserRepository, db *sql.DB) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		db:       db,
	}
}

func (s *AuthService) GetOrCreateUserFromKeycloak(ctx context.Context, claims *keycloak.Claims) (*models.User, error) {
	log.Printf("[AuthService] GetOrCreateUserFromKeycloak: KeycloakID=%s, Email=%s, Username=%s",
		claims.Subject, claims.Email, claims.PreferredUsername)

	// First, try to find by keycloak_id
	user, err := s.userRepo.GetByKeycloakID(ctx, claims.Subject)
	if err == nil {
		log.Printf("[AuthService] Found existing user by KeycloakID: ID=%d", user.ID)
		if err := s.userRepo.UpdateStatus(ctx, user.ID, models.UserStatusOnline); err != nil {
			log.Printf("[AuthService] Failed to update status: %v", err)
			return nil, err
		}
		user.Status = models.UserStatusOnline
		return user, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		log.Printf("[AuthService] GetByKeycloakID error: %v", err)
		return nil, err
	}

	// Not found by keycloak_id, try to find by email
	log.Printf("[AuthService] User not found by KeycloakID, trying email lookup")
	user, err = s.userRepo.GetByEmail(ctx, claims.Email)
	if err == nil {
		// Found existing user by email, update keycloak_id
		log.Printf("[AuthService] Found existing user by email: ID=%d, updating KeycloakID", user.ID)
		if err := s.userRepo.UpdateKeycloakID(ctx, user.ID, claims.Subject); err != nil {
			log.Printf("[AuthService] Failed to update KeycloakID: %v", err)
			return nil, err
		}
		if err := s.userRepo.UpdateStatus(ctx, user.ID, models.UserStatusOnline); err != nil {
			log.Printf("[AuthService] Failed to update status: %v", err)
			return nil, err
		}
		user.KeycloakID = sql.NullString{String: claims.Subject, Valid: true}
		user.Status = models.UserStatusOnline
		return user, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		log.Printf("[AuthService] GetByEmail error: %v", err)
		return nil, err
	}

	// User not found, create new user
	log.Printf("[AuthService] User not found, creating new user")

	username := claims.PreferredUsername
	if username == "" {
		username = claims.Email
	}

	user = &models.User{
		KeycloakID: sql.NullString{String: claims.Subject, Valid: true},
		Email:      claims.Email,
		Username:   username,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Printf("[AuthService] Failed to create user: %v", err)
		return nil, err
	}

	log.Printf("[AuthService] Created new user: ID=%d", user.ID)

	if err := s.userRepo.UpdateStatus(ctx, user.ID, models.UserStatusOnline); err != nil {
		log.Printf("[AuthService] Failed to update status for new user: %v", err)
		return nil, err
	}
	user.Status = models.UserStatusOnline

	return user, nil
}

func (s *AuthService) GetUserByKeycloakID(ctx context.Context, keycloakID string) (*models.User, error) {
	return s.userRepo.GetByKeycloakID(ctx, keycloakID)
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uint64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) Logout(ctx context.Context, userID uint64) error {
	return s.userRepo.UpdateStatus(ctx, userID, models.UserStatusOffline)
}
