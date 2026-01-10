package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"Mmessenger/internal/models"
	"Mmessenger/internal/repository"
	"Mmessenger/pkg/jwt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtService *jwt.Service
	db         *sql.DB
}

func NewAuthService(userRepo *repository.UserRepository, jwtService *jwt.Service, db *sql.DB) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
		db:         db,
	}
}

func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.AuthResponse, error) {
	// Check if user already exists
	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Store refresh token hash
	if err := s.storeRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update user status to online
	if err := s.userRepo.UpdateStatus(ctx, user.ID, models.UserStatusOnline); err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Store refresh token hash
	if err := s.storeRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	user.Status = models.UserStatusOnline
	return &models.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uint64, refreshToken string) error {
	// Revoke refresh token
	tokenHash := s.hashToken(refreshToken)
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = ? AND user_id = ?`
	_, err := s.db.ExecContext(ctx, query, tokenHash, userID)
	if err != nil {
		return err
	}

	// Update user status to offline
	return s.userRepo.UpdateStatus(ctx, userID, models.UserStatusOffline)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.AuthResponse, error) {
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Verify refresh token in database
	tokenHash := s.hashToken(refreshToken)
	var userID uint64
	query := `SELECT user_id FROM refresh_tokens WHERE token_hash = ? AND expires_at > NOW() AND revoked_at IS NULL`
	err = s.db.QueryRowContext(ctx, query, tokenHash).Scan(&userID)
	if err != nil {
		return nil, jwt.ErrInvalidToken
	}

	if userID != claims.UserID {
		return nil, jwt.ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:        user.ToResponse(),
		AccessToken: accessToken,
	}, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uint64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) storeRefreshToken(ctx context.Context, userID uint64, token string, expiresAt interface{}) error {
	tokenHash := s.hashToken(token)
	query := `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, userID, tokenHash, expiresAt)
	return err
}

func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
