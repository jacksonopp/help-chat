package handlers

import (
	"net/http"
	"time"

	authMiddleware "dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/middleware"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/services"

	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRoutes registers all authentication-related routes
func (h *AuthHandler) RegisterRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
	// API v1 routes
	api := e.Group("/api/v1")

	// Authentication routes
	auth := api.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.RefreshToken)
	auth.POST("/logout", h.Logout, authMiddlewareInstance.Authenticate)
	auth.POST("/forgot-password", h.ForgotPassword)
	auth.POST("/reset-password", h.ResetPassword)
	auth.POST("/verify-email", h.VerifyEmail)
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account with the specified role
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration request"
// @Success 201 {object} models.AuthResponse "User registered successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request data"
// @Failure 409 {object} models.ErrorResponse "User already exists"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Register user
	response, tokenResponse, err := h.authService.Register(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Set JWT tokens as HTTP-only cookies
	h.setAuthCookies(c, tokenResponse.AccessToken, tokenResponse.RefreshToken)

	// Return response without tokens
	return c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT tokens as cookies
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} models.AuthResponse "Login successful"
// @Failure 400 {object} models.ErrorResponse "Invalid request data"
// @Failure 401 {object} models.ErrorResponse "Invalid credentials"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Login user
	response, tokenResponse, err := h.authService.Login(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	// Set JWT tokens as HTTP-only cookies
	h.setAuthCookies(c, tokenResponse.AccessToken, tokenResponse.RefreshToken)

	// Return response without tokens
	return c.JSON(http.StatusOK, response)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Generate new access token using refresh token from cookie
// @Tags authentication
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse "Token refreshed successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request data"
// @Failure 401 {object} models.ErrorResponse "Invalid refresh token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	// Get refresh token from cookie
	refreshTokenCookie, err := c.Cookie("refresh_token")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh token not found")
	}

	// Refresh token
	response, err := h.authService.RefreshToken(refreshTokenCookie.Value)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	// Set new access token as cookie
	accessTokenTTL, err := time.ParseDuration(h.authService.GetConfig().JWT.AccessTokenTTL)
	if err != nil {
		accessTokenTTL = 15 * time.Minute // fallback
	}

	// Determine SameSite value
	var sameSite http.SameSite
	switch h.authService.GetConfig().JWT.CookieSameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "Lax":
		sameSite = http.SameSiteLaxMode
	case "None":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteLaxMode
	}

	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    response.AccessToken,
		Path:     "/",
		Domain:   h.authService.GetConfig().JWT.CookieDomain,
		Expires:  time.Now().Add(accessTokenTTL),
		HttpOnly: true,
		Secure:   h.authService.GetConfig().JWT.CookieSecure,
		SameSite: sameSite,
	})

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Token refreshed successfully",
	})
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and clear authentication cookies
// @Tags authentication
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.SuccessResponse "Logout successful"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	// Determine SameSite value
	var sameSite http.SameSite
	switch h.authService.GetConfig().JWT.CookieSameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "Lax":
		sameSite = http.SameSiteLaxMode
	case "None":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteLaxMode
	}

	// Clear authentication cookies
	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Domain:   h.authService.GetConfig().JWT.CookieDomain,
		HttpOnly: true,
		Secure:   h.authService.GetConfig().JWT.CookieSecure,
		SameSite: sameSite,
		MaxAge:   -1, // Delete the cookie
	})

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Domain:   h.authService.GetConfig().JWT.CookieDomain,
		HttpOnly: true,
		Secure:   h.authService.GetConfig().JWT.CookieSecure,
		SameSite: sameSite,
		MaxAge:   -1, // Delete the cookie
	})

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Logout successful",
	})
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} models.SuccessResponse "Password reset email sent"
// @Failure 400 {object} models.ErrorResponse "Invalid request data"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req models.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// TODO: Implement password reset email functionality
	// For now, we'll just return success
	return c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password using reset token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} models.SuccessResponse "Password reset successful"
// @Failure 400 {object} models.ErrorResponse "Invalid request data"
// @Failure 401 {object} models.ErrorResponse "Invalid or expired token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req models.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// TODO: Implement password reset functionality
	// For now, we'll just return success
	return c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Password reset successful",
	})
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify user email address using verification token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.VerifyEmailRequest true "Email verification request"
// @Success 200 {object} models.SuccessResponse "Email verified successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request data"
// @Failure 401 {object} models.ErrorResponse "Invalid or expired token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	var req models.VerifyEmailRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// TODO: Implement email verification functionality
	// For now, we'll just return success
	return c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Email verified successfully",
	})
}

func (h *AuthHandler) setAuthCookies(c echo.Context, accessToken, refreshToken string) {
	// Parse access token TTL for cookie expiration
	accessTokenTTL, err := time.ParseDuration(h.authService.GetConfig().JWT.AccessTokenTTL)
	if err != nil {
		accessTokenTTL = 15 * time.Minute // fallback
	}

	// Parse refresh token TTL for cookie expiration
	refreshTokenTTL, err := time.ParseDuration(h.authService.GetConfig().JWT.RefreshTokenTTL)
	if err != nil {
		refreshTokenTTL = 7 * 24 * time.Hour // fallback
	}

	// Determine SameSite value
	var sameSite http.SameSite
	switch h.authService.GetConfig().JWT.CookieSameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "Lax":
		sameSite = http.SameSiteLaxMode
	case "None":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteLaxMode
	}

	// Set access token cookie
	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    accessToken,
		Path:     "/",
		Domain:   h.authService.GetConfig().JWT.CookieDomain,
		Expires:  time.Now().Add(accessTokenTTL),
		HttpOnly: true,
		Secure:   h.authService.GetConfig().JWT.CookieSecure,
		SameSite: sameSite,
	})

	// Set refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   h.authService.GetConfig().JWT.CookieDomain,
		Expires:  time.Now().Add(refreshTokenTTL),
		HttpOnly: true,
		Secure:   h.authService.GetConfig().JWT.CookieSecure,
		SameSite: sameSite,
	})
}
