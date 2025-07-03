package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/config"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/handlers"
	testMiddleware "dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/middleware"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/repository"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/services"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/pkg/database"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestAuthEndpoints tests the authentication endpoints
func TestAuthEndpoints(t *testing.T) {
	// Setup test database
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			FilePath: ":memory:", // Use in-memory database for testing
		},
		JWT: config.JWTConfig{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  "15m",
			RefreshTokenTTL: "7d",
			Issuer:          "test",
		},
	}

	db, err := database.NewDatabase(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Run migrations
	err = database.RunMigrations(db)
	assert.NoError(t, err)

	// Initialize components
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, cfg)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Echo with validator
	e := echo.New()
	e.Validator = testMiddleware.NewCustomValidator()

	// Test registration
	t.Run("Register", func(t *testing.T) {
		registerReq := models.RegisterRequest{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			Role:      models.RoleEndUser,
		}

		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Test registration
		err := authHandler.Register(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Verify response
		var response models.AuthResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", response.User.Email)

		// Verify cookies are set
		cookies := rec.Result().Cookies()
		var tokenCookie, refreshCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "token" {
				tokenCookie = cookie
			}
			if cookie.Name == "refresh_token" {
				refreshCookie = cookie
			}
		}
		assert.NotNil(t, tokenCookie, "Access token cookie should be set")
		assert.NotNil(t, refreshCookie, "Refresh token cookie should be set")
		assert.NotEmpty(t, tokenCookie.Value, "Access token should not be empty")
		assert.NotEmpty(t, refreshCookie.Value, "Refresh token should not be empty")
		assert.True(t, tokenCookie.HttpOnly, "Token cookie should be HttpOnly")
		assert.True(t, refreshCookie.HttpOnly, "Refresh token cookie should be HttpOnly")
	})

	// Test login
	t.Run("Login", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		reqBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Test login
		err := authHandler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify response
		var response models.AuthResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", response.User.Email)

		// Verify cookies are set
		cookies := rec.Result().Cookies()
		var tokenCookie, refreshCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "token" {
				tokenCookie = cookie
			}
			if cookie.Name == "refresh_token" {
				refreshCookie = cookie
			}
		}
		assert.NotNil(t, tokenCookie, "Access token cookie should be set")
		assert.NotNil(t, refreshCookie, "Refresh token cookie should be set")
		assert.NotEmpty(t, tokenCookie.Value, "Access token should not be empty")
		assert.NotEmpty(t, refreshCookie.Value, "Refresh token should not be empty")
		assert.True(t, tokenCookie.HttpOnly, "Token cookie should be HttpOnly")
		assert.True(t, refreshCookie.HttpOnly, "Refresh token cookie should be HttpOnly")
	})

	// Test invalid login
	t.Run("InvalidLogin", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		reqBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Test invalid login
		err := authHandler.Login(c)
		assert.Error(t, err)
		// Echo returns an HTTP error, so we need to check the error type
		httpError, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, httpError.Code)
	})
}
