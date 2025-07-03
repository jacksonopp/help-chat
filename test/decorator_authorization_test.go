package test

import (
	"net/http/httptest"
	"testing"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/handlers"
	authMiddleware "dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/middleware"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTicketService for testing
type MockTicketService struct {
	mock.Mock
}

func (m *MockTicketService) CreateTicket(ctx echo.Context, req *models.CreateTicketRequest, userID uuid.UUID) (*models.Ticket, error) {
	args := m.Called(ctx, req, userID)
	return args.Get(0).(*models.Ticket), args.Error(1)
}

func (m *MockTicketService) GetTicket(ctx echo.Context, ticketID uuid.UUID) (*models.Ticket, error) {
	args := m.Called(ctx, ticketID)
	return args.Get(0).(*models.Ticket), args.Error(1)
}

func (m *MockTicketService) UpdateTicket(ctx echo.Context, ticketID uuid.UUID, req *models.UpdateTicketRequest, userID uuid.UUID) (*models.Ticket, error) {
	args := m.Called(ctx, ticketID, req, userID)
	return args.Get(0).(*models.Ticket), args.Error(1)
}

func (m *MockTicketService) DeleteTicket(ctx echo.Context, ticketID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, ticketID, userID)
	return args.Error(0)
}

func (m *MockTicketService) ListTickets(ctx echo.Context, query *models.TicketQuery) (*models.TicketListResponse, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*models.TicketListResponse), args.Error(1)
}

func (m *MockTicketService) AssignTicket(ctx echo.Context, ticketID uuid.UUID, agentID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, ticketID, agentID, userID)
	return args.Error(0)
}

func (m *MockTicketService) UpdateTicketStatus(ctx echo.Context, ticketID uuid.UUID, req *models.UpdateTicketStatusRequest, userID uuid.UUID) error {
	args := m.Called(ctx, ticketID, req, userID)
	return args.Error(0)
}

func (m *MockTicketService) EscalateTicket(ctx echo.Context, ticketID uuid.UUID, req *models.EscalateTicketRequest, userID uuid.UUID) error {
	args := m.Called(ctx, ticketID, req, userID)
	return args.Error(0)
}

func (m *MockTicketService) GetTicketsByUser(ctx echo.Context, userID uuid.UUID, query *models.TicketQuery) (*models.TicketListResponse, error) {
	args := m.Called(ctx, userID, query)
	return args.Get(0).(*models.TicketListResponse), args.Error(1)
}

func (m *MockTicketService) GetTicketsByAgent(ctx echo.Context, agentID uuid.UUID, query *models.TicketQuery) (*models.TicketListResponse, error) {
	args := m.Called(ctx, agentID, query)
	return args.Get(0).(*models.TicketListResponse), args.Error(1)
}

func (m *MockTicketService) GetTicketStats(ctx echo.Context) (*models.TicketStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.TicketStats), args.Error(1)
}

func TestDecoratorAuthorization(t *testing.T) {
	// Setup
	e := echo.New()
	mockTicketService := new(MockTicketService)
	authMiddlewareInstance := &authMiddleware.AuthMiddleware{}

	// Create handlers
	decoratedHandler := handlers.NewDecoratedTicketHandler(mockTicketService)
	taggedHandler := handlers.NewTaggedTicketHandler(mockTicketService)

	// Register routes
	decoratedHandler.RegisterDecoratedRoutes(e, authMiddlewareInstance)
	taggedHandler.RegisterTaggedRoutes(e, authMiddlewareInstance)

	// Test cases
	testCases := []struct {
		name           string
		method         string
		path           string
		user           *models.User
		expectedStatus int
		description    string
	}{
		{
			name:           "Admin can access admin-only decorated route",
			method:         "DELETE",
			path:           "/api/v1/decorated-tickets/test-id",
			user:           createTestUser(models.RoleAdministrator),
			expectedStatus: 204,
			description:    "Administrator should be able to access admin-only route",
		},
		{
			name:           "Manager can access admin-only decorated route",
			method:         "DELETE",
			path:           "/api/v1/decorated-tickets/test-id",
			user:           createTestUser(models.RoleManager),
			expectedStatus: 204,
			description:    "Manager should be able to access admin-only route",
		},
		{
			name:           "Agent cannot access admin-only decorated route",
			method:         "DELETE",
			path:           "/api/v1/decorated-tickets/test-id",
			user:           createTestUser(models.RoleSupportAgent),
			expectedStatus: 403,
			description:    "Agent should not be able to access admin-only route",
		},
		{
			name:           "End user cannot access admin-only decorated route",
			method:         "DELETE",
			path:           "/api/v1/decorated-tickets/test-id",
			user:           createTestUser(models.RoleEndUser),
			expectedStatus: 403,
			description:    "End user should not be able to access admin-only route",
		},
		{
			name:           "Admin can access admin-only tagged route",
			method:         "DELETE",
			path:           "/api/v1/tagged-tickets/test-id",
			user:           createTestUser(models.RoleAdministrator),
			expectedStatus: 204,
			description:    "Administrator should be able to access admin-only tagged route",
		},
		{
			name:           "Manager can access admin-only tagged route",
			method:         "DELETE",
			path:           "/api/v1/tagged-tickets/test-id",
			user:           createTestUser(models.RoleManager),
			expectedStatus: 204,
			description:    "Manager should be able to access admin-only tagged route",
		},
		{
			name:           "Agent cannot access admin-only tagged route",
			method:         "DELETE",
			path:           "/api/v1/tagged-tickets/test-id",
			user:           createTestUser(models.RoleSupportAgent),
			expectedStatus: 403,
			description:    "Agent should not be able to access admin-only tagged route",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock expectations
			if tc.expectedStatus == 204 {
				mockTicketService.On("DeleteTicket", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}

			// Create request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set user in context (simulating authenticated request)
			c.Set("user", tc.user)
			c.Set("user_id", tc.user.ID.String())
			c.Set("user_role", string(tc.user.Role))

			// Execute request
			e.ServeHTTP(rec, req)

			// Assertions
			assert.Equal(t, tc.expectedStatus, rec.Code, tc.description)

			// Verify mock expectations
			mockTicketService.AssertExpectations(t)
		})
	}
}

// Helper function to create test users
func createTestUser(role models.UserRole) *models.User {
	return &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  role,
	}
}

// Test the decorator system directly
func TestDecoratorSystem(t *testing.T) {
	authMiddlewareInstance := &authMiddleware.AuthMiddleware{}

	// Test AuthorizeAdmin decorator
	t.Run("AuthorizeAdmin decorator", func(t *testing.T) {
		decorator := authMiddleware.AuthorizeAdmin()
		assert.NotNil(t, decorator)
		assert.Equal(t, "Requires admin privileges", decorator.Description)
		assert.Contains(t, decorator.Roles, models.RoleAdministrator)
		assert.Contains(t, decorator.Roles, models.RoleManager)
	})

	// Test AuthorizeAgent decorator
	t.Run("AuthorizeAgent decorator", func(t *testing.T) {
		decorator := authMiddleware.AuthorizeAgent()
		assert.NotNil(t, decorator)
		assert.Equal(t, "Requires agent privileges", decorator.Description)
		assert.Contains(t, decorator.Roles, models.RoleSupportAgent)
		assert.Contains(t, decorator.Roles, models.RoleAdministrator)
		assert.Contains(t, decorator.Roles, models.RoleManager)
	})

	// Test AuthorizePermission decorator
	t.Run("AuthorizePermission decorator", func(t *testing.T) {
		decorator := authMiddleware.AuthorizePermission("ticket:delete", "ticket:admin")
		assert.NotNil(t, decorator)
		assert.Equal(t, "Requires specific permissions", decorator.Description)
		assert.Contains(t, decorator.Permissions, "ticket:delete")
		assert.Contains(t, decorator.Permissions, "ticket:admin")
	})

	// Test AuthorizeAny decorator
	t.Run("AuthorizeAny decorator", func(t *testing.T) {
		decorator := authMiddleware.AuthorizeAny(
			[]models.UserRole{models.RoleAdministrator, models.RoleManager},
			[]string{"ticket:update", "ticket:admin"},
		)
		assert.NotNil(t, decorator)
		assert.Equal(t, "Requires any of the specified roles or permissions", decorator.Description)
		assert.Contains(t, decorator.Roles, models.RoleAdministrator)
		assert.Contains(t, decorator.Roles, models.RoleManager)
		assert.Contains(t, decorator.Permissions, "ticket:update")
		assert.Contains(t, decorator.Permissions, "ticket:admin")
	})
}

// Test the tag parsing system
func TestTagParsing(t *testing.T) {
	// Test role parsing
	t.Run("Parse role tags", func(t *testing.T) {
		tag := authMiddleware.ParseAuthorizeTag("ADMIN,MANAGER")
		assert.NotNil(t, tag)
		assert.Contains(t, tag.Roles, models.RoleAdministrator)
		assert.Contains(t, tag.Roles, models.RoleManager)
	})

	// Test permission parsing
	t.Run("Parse permission tags", func(t *testing.T) {
		tag := authMiddleware.ParseAuthorizeTag("ticket:delete,ticket:admin")
		assert.NotNil(t, tag)
		assert.Contains(t, tag.Permissions, "ticket:delete")
		assert.Contains(t, tag.Permissions, "ticket:admin")
	})

	// Test mixed parsing
	t.Run("Parse mixed tags", func(t *testing.T) {
		tag := authMiddleware.ParseAuthorizeTag("ADMIN,MANAGER,ticket:update")
		assert.NotNil(t, tag)
		assert.Contains(t, tag.Roles, models.RoleAdministrator)
		assert.Contains(t, tag.Roles, models.RoleManager)
		assert.Contains(t, tag.Permissions, "ticket:update")
	})

	// Test policy parsing
	t.Run("Parse policy tags", func(t *testing.T) {
		tag := authMiddleware.ParseAuthorizeTag("Policy=ticket:status:update")
		assert.NotNil(t, tag)
		assert.Equal(t, "ticket:status:update", tag.Policy)
	})

	// Test empty tag
	t.Run("Parse empty tag", func(t *testing.T) {
		tag := authMiddleware.ParseAuthorizeTag("")
		assert.Nil(t, tag)
	})
}
