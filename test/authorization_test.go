package test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/config"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/handlers"
	authMiddleware "dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/middleware"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/repository"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/services"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/pkg/database"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizationMiddleware(t *testing.T) {
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
	ticketRepo := repository.NewTicketRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	attachmentRepo := repository.NewAttachmentRepository(db)

	authService := services.NewAuthService(userRepo, cfg)
	ticketService := services.NewTicketService(ticketRepo, categoryRepo, commentRepo, attachmentRepo, userRepo)

	authHandler := handlers.NewAuthHandler(authService)
	ticketHandler := handlers.NewTicketHandler(ticketService)
	authMiddlewareInstance := authMiddleware.NewAuthMiddleware(authService)

	// Setup Echo
	e := echo.New()
	e.Validator = authMiddleware.NewCustomValidator()

	// Register routes
	authHandler.RegisterRoutes(e, authMiddlewareInstance)
	ticketHandler.RegisterRoutes(e, authMiddlewareInstance)

	// Create test users
	adminUser := createTestUser(t, userRepo, "admin@test.com", "password123", models.RoleAdministrator)
	agentUser := createTestUser(t, userRepo, "agent@test.com", "password123", models.RoleSupportAgent)
	endUser := createTestUser(t, userRepo, "user@test.com", "password123", models.RoleEndUser)

	// Test cases
	testCases := []struct {
		name           string
		user           *models.User
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Admin can delete ticket",
			user:           adminUser,
			method:         "DELETE",
			path:           "/api/v1/tickets/test-id",
			expectedStatus: http.StatusInternalServerError, // Will fail due to invalid ticket ID, but not due to auth
			description:    "Admin should be able to delete tickets",
		},
		{
			name:           "Agent cannot delete ticket",
			user:           agentUser,
			method:         "DELETE",
			path:           "/api/v1/tickets/test-id",
			expectedStatus: http.StatusForbidden,
			description:    "Agent should not be able to delete tickets",
		},
		{
			name:           "End user cannot delete ticket",
			user:           endUser,
			method:         "DELETE",
			path:           "/api/v1/tickets/test-id",
			expectedStatus: http.StatusForbidden,
			description:    "End user should not be able to delete tickets",
		},
		{
			name:           "Agent can assign ticket",
			user:           agentUser,
			method:         "POST",
			path:           "/api/v1/tickets/test-id/assign",
			expectedStatus: http.StatusInternalServerError, // Will fail due to invalid ticket ID, but not due to auth
			description:    "Agent should be able to assign tickets",
		},
		{
			name:           "End user cannot assign ticket",
			user:           endUser,
			method:         "POST",
			path:           "/api/v1/tickets/test-id/assign",
			expectedStatus: http.StatusForbidden,
			description:    "End user should not be able to assign tickets",
		},
		{
			name:           "Agent can view stats",
			user:           agentUser,
			method:         "GET",
			path:           "/api/v1/tickets/stats",
			expectedStatus: http.StatusOK,
			description:    "Agent should be able to view ticket stats",
		},
		{
			name:           "End user cannot view stats",
			user:           endUser,
			method:         "GET",
			path:           "/api/v1/tickets/stats",
			expectedStatus: http.StatusForbidden,
			description:    "End user should not be able to view ticket stats",
		},
		{
			name:           "Any authenticated user can create ticket",
			user:           endUser,
			method:         "POST",
			path:           "/api/v1/tickets",
			expectedStatus: http.StatusBadRequest, // Will fail due to invalid request body, but not due to auth
			description:    "Any authenticated user should be able to create tickets",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			var req *http.Request
			if tc.method == "POST" {
				reqBody := []byte(`{"test": "data"}`)
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewReader(reqBody))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set user in context (simulating authenticated request)
			c.Set("user", tc.user)
			c.Set("user_id", tc.user.ID.String())
			c.Set("user_role", string(tc.user.Role))

			// Execute request using Echo's ServeHTTP method
			e.ServeHTTP(rec, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rec.Code, tc.description)
		})
	}
}

func TestPermissionSystem(t *testing.T) {
	// Test the permission system directly
	authMiddlewareInstance := &authMiddleware.AuthMiddleware{}

	// Test role permissions
	testCases := []struct {
		role        models.UserRole
		permission  string
		expected    bool
		description string
	}{
		{
			role:        models.RoleAdministrator,
			permission:  "ticket:delete",
			expected:    true,
			description: "Administrator should have ticket:delete permission",
		},
		{
			role:        models.RoleManager,
			permission:  "ticket:delete",
			expected:    true,
			description: "Manager should have ticket:delete permission",
		},
		{
			role:        models.RoleSupportAgent,
			permission:  "ticket:delete",
			expected:    false,
			description: "Support agent should not have ticket:delete permission",
		},
		{
			role:        models.RoleEndUser,
			permission:  "ticket:delete",
			expected:    false,
			description: "End user should not have ticket:delete permission",
		},
		{
			role:        models.RoleSupportAgent,
			permission:  "ticket:assign",
			expected:    true,
			description: "Support agent should have ticket:assign permission",
		},
		{
			role:        models.RoleEndUser,
			permission:  "ticket:assign",
			expected:    false,
			description: "End user should not have ticket:assign permission",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			hasPermission := authMiddlewareInstance.HasPermission(tc.role, tc.permission)
			assert.Equal(t, tc.expected, hasPermission, tc.description)
		})
	}
}

// Helper function to create test users
func createTestUser(t *testing.T, userRepo repository.UserRepository, email, password string, role models.UserRole) *models.User {
	user := &models.User{
		Email:        email,
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
		FirstName:    "Test",
		LastName:     "User",
		Role:         role,
		IsVerified:   true,
		IsActive:     true,
	}

	err := userRepo.Create(user)
	assert.NoError(t, err)

	return user
}
