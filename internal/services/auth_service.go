package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/config"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication-related operations
type AuthService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, config *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   config,
	}
}

// Register creates a new user account
func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, *models.TokenResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         req.Role,
		IsVerified:   false,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokenResponse, err := s.generateTokens(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &models.AuthResponse{
		User: user,
	}, tokenResponse, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, *models.TokenResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, fmt.Errorf("account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(user); err != nil {
		return nil, nil, fmt.Errorf("failed to update last login time: %w", err)
	}

	// Generate tokens
	tokenResponse, err := s.generateTokens(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &models.AuthResponse{
		User: user,
	}, tokenResponse, nil
}

// RefreshToken generates new access token using refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*models.TokenResponse, error) {
	// Parse and validate refresh token
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if token is a refresh token
	if claims["token_type"] != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Get user ID from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	// Get user
	user, err := s.userRepo.GetByID(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Parse TTL for response
	accessTokenTTL, err := time.ParseDuration(s.config.JWT.AccessTokenTTL)
	if err != nil {
		accessTokenTTL = 15 * time.Minute // fallback
	}

	return &models.TokenResponse{
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(accessTokenTTL),
		TokenType:   "Bearer",
	}, nil
}

// generateTokens generates both access and refresh tokens
func (s *AuthService) generateTokens(user *models.User) (*models.TokenResponse, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	// Parse TTL for response
	accessTokenTTL, err := time.ParseDuration(s.config.JWT.AccessTokenTTL)
	if err != nil {
		accessTokenTTL = 15 * time.Minute // fallback
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(accessTokenTTL),
		TokenType:    "Bearer",
	}, nil
}

// generateAccessToken generates an access token
func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	accessTokenTTL, err := time.ParseDuration(s.config.JWT.AccessTokenTTL)
	if err != nil {
		accessTokenTTL = 15 * time.Minute // fallback
	}

	claims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"role":       string(user.Role),
		"token_type": "access",
		"exp":        time.Now().Add(accessTokenTTL).Unix(),
		"iat":        time.Now().Unix(),
		"iss":        s.config.JWT.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.SecretKey))
}

// generateRefreshToken generates a refresh token
func (s *AuthService) generateRefreshToken(user *models.User) (string, error) {
	refreshTokenTTL, err := time.ParseDuration(s.config.JWT.RefreshTokenTTL)
	if err != nil {
		refreshTokenTTL = 7 * 24 * time.Hour // fallback
	}

	claims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"role":       string(user.Role),
		"token_type": "refresh",
		"exp":        time.Now().Add(refreshTokenTTL).Unix(),
		"iat":        time.Now().Unix(),
		"iss":        s.config.JWT.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.SecretKey))
}

// parseToken parses and validates a JWT token
func (s *AuthService) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateToken validates an access token and returns user claims
func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if token is an access token
	if claims["token_type"] != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Get user ID from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	// Get user from database
	user, err := s.userRepo.GetByID(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	return user, nil
}

// generateRandomToken generates a random token for password reset and email verification
func (s *AuthService) generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetConfig returns the configuration
func (s *AuthService) GetConfig() *config.Config {
	return s.config
}
