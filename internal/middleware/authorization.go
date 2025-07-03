package middleware

import (
	"github.com/labstack/echo/v4"
)

// RoutePermission defines the permission required for a specific route
type RoutePermission struct {
	Method      string
	Path        string
	Permission  string
	Description string
}

// AuthorizationConfig defines the authorization configuration for the application
type AuthorizationConfig struct {
	RoutePermissions []RoutePermission
}

// NewAuthorizationConfig creates a new authorization configuration
func NewAuthorizationConfig() *AuthorizationConfig {
	return &AuthorizationConfig{
		RoutePermissions: []RoutePermission{
			// Ticket routes
			{
				Method:      "GET",
				Path:        "/api/v1/tickets",
				Permission:  "ticket:read",
				Description: "List all tickets",
			},
			{
				Method:      "POST",
				Path:        "/api/v1/tickets",
				Permission:  "ticket:create",
				Description: "Create a new ticket",
			},
			{
				Method:      "GET",
				Path:        "/api/v1/tickets/:id",
				Permission:  "ticket:read",
				Description: "Get a specific ticket",
			},
			{
				Method:      "PUT",
				Path:        "/api/v1/tickets/:id",
				Permission:  "ticket:update",
				Description: "Update a ticket",
			},
			{
				Method:      "DELETE",
				Path:        "/api/v1/tickets/:id",
				Permission:  "ticket:delete",
				Description: "Delete a ticket (admin only)",
			},
			{
				Method:      "POST",
				Path:        "/api/v1/tickets/:id/assign",
				Permission:  "ticket:assign",
				Description: "Assign a ticket to an agent",
			},
			{
				Method:      "POST",
				Path:        "/api/v1/tickets/:id/status",
				Permission:  "ticket:status:update",
				Description: "Update ticket status",
			},
			{
				Method:      "POST",
				Path:        "/api/v1/tickets/:id/escalate",
				Permission:  "ticket:escalate",
				Description: "Escalate a ticket",
			},
			{
				Method:      "GET",
				Path:        "/api/v1/tickets/my",
				Permission:  "ticket:read:own",
				Description: "Get user's own tickets",
			},
			{
				Method:      "GET",
				Path:        "/api/v1/tickets/assigned",
				Permission:  "ticket:read",
				Description: "Get tickets assigned to user",
			},
			{
				Method:      "GET",
				Path:        "/api/v1/tickets/stats",
				Permission:  "ticket:stats:read",
				Description: "Get ticket statistics",
			},
		},
	}
}

// GetPermissionForRoute returns the permission required for a specific route
func (ac *AuthorizationConfig) GetPermissionForRoute(method, path string) (string, bool) {
	for _, route := range ac.RoutePermissions {
		if route.Method == method && route.Path == path {
			return route.Permission, true
		}
	}
	return "", false
}

// RequireRoutePermission creates middleware that requires permission for a specific route
func (m *AuthMiddleware) RequireRoutePermission(config *AuthorizationConfig, method, path string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			permission, exists := config.GetPermissionForRoute(method, path)
			if !exists {
				// If no specific permission is defined, allow access
				return next(c)
			}

			return m.RequirePermission(permission)(next)(c)
		}
	}
}

// AuthorizationMiddleware creates middleware that automatically checks permissions based on route
func (m *AuthMiddleware) AuthorizationMiddleware(config *AuthorizationConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			path := c.Path()

			permission, exists := config.GetPermissionForRoute(method, path)
			if !exists {
				// If no specific permission is defined, allow access
				return next(c)
			}

			return m.RequirePermission(permission)(next)(c)
		}
	}
}
