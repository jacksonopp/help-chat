package middleware

import (
	"net/http"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/services"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	authService *services.AuthService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate validates JWT tokens and sets user context
func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get token from cookie
		tokenCookie, err := c.Cookie("token")
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing authentication token")
		}

		// Validate token
		user, err := m.authService.ValidateToken(tokenCookie.Value)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		// Set user in context
		c.Set("user", user)
		c.Set("user_id", user.ID.String())
		c.Set("user_role", string(user.Role))

		return next(c)
	}
}

// RequireRole creates middleware that requires a specific user role
func (m *AuthMiddleware) RequireRole(requiredRole models.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
			}

			if user.Role != requiredRole {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(c)
		}
	}
}

// RequireAnyRole creates middleware that requires any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(requiredRoles ...models.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
			}

			hasRole := false
			for _, role := range requiredRoles {
				if user.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(c)
		}
	}
}

// RequireAdmin creates middleware that requires admin privileges
func (m *AuthMiddleware) RequireAdmin() echo.MiddlewareFunc {
	return m.RequireAnyRole(models.RoleAdministrator, models.RoleManager)
}

// RequireAgent creates middleware that requires agent privileges
func (m *AuthMiddleware) RequireAgent() echo.MiddlewareFunc {
	return m.RequireAnyRole(models.RoleSupportAgent, models.RoleAdministrator, models.RoleManager)
}

// RequireManager creates middleware that requires manager or admin privileges
func (m *AuthMiddleware) RequireManager() echo.MiddlewareFunc {
	return m.RequireAnyRole(models.RoleManager, models.RoleAdministrator)
}

// RequireSpecificRole creates middleware that requires a specific role
func (m *AuthMiddleware) RequireSpecificRole(role models.UserRole) echo.MiddlewareFunc {
	return m.RequireRole(role)
}

type OwnerIdGetter func(c echo.Context) (string, error)

// RequireOwnerOrAdmin creates middleware that allows access if user owns the resource or is admin
func (m *AuthMiddleware) RequireOwnerOrAdmin(ownerIDGetter OwnerIdGetter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
			}

			// Admin can access anything
			if user.IsAdmin() {
				return next(c)
			}

			// Get the owner ID of the resource
			ownerID, err := ownerIDGetter(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "unable to determine resource ownership")
			}

			// Check if user is the owner
			if user.ID.String() == ownerID {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

// RequirePermission creates middleware that checks for specific permissions
func (m *AuthMiddleware) RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
			}

			// For now, we'll use role-based permissions
			// In a more complex system, you might have a separate permissions table
			hasPermission := m.HasPermission(user.Role, permission)
			if !hasPermission {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(c)
		}
	}
}

// HasPermission checks if a role has a specific permission
func (m *AuthMiddleware) HasPermission(role models.UserRole, permission string) bool {
	// Define role-based permissions
	permissions := map[models.UserRole][]string{
		models.RoleEndUser: {
			"ticket:create",
			"ticket:read:own",
			"ticket:update:own",
		},
		models.RoleSupportAgent: {
			"ticket:create",
			"ticket:read",
			"ticket:update",
			"ticket:assign",
			"ticket:status:update",
			"ticket:escalate",
			"ticket:stats:read",
		},
		models.RoleManager: {
			"ticket:create",
			"ticket:read",
			"ticket:update",
			"ticket:assign",
			"ticket:status:update",
			"ticket:escalate",
			"ticket:stats:read",
			"ticket:delete",
			"user:manage",
		},
		models.RoleAdministrator: {
			"ticket:create",
			"ticket:read",
			"ticket:update",
			"ticket:assign",
			"ticket:status:update",
			"ticket:escalate",
			"ticket:stats:read",
			"ticket:delete",
			"user:manage",
			"system:admin",
		},
	}

	rolePermissions, exists := permissions[role]
	if !exists {
		return false
	}

	for _, perm := range rolePermissions {
		if perm == permission {
			return true
		}
	}

	return false
}
