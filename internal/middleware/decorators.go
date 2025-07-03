package middleware

import (
	"strings"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"github.com/labstack/echo/v4"
)

// AuthorizationDecorator defines authorization requirements for a handler method
type AuthorizationDecorator struct {
	Roles       []models.UserRole
	Permissions []string
	Description string
}

// AuthorizeAll decorator for allowing all authenticated users
func AuthorizeAll() *AuthorizationDecorator {
	return &AuthorizationDecorator{
		Roles:       []models.UserRole{models.RoleEndUser, models.RoleSupportAgent, models.RoleAdministrator, models.RoleManager},
		Description: "Allows all authenticated users",
	}
}

// Authorize decorator for requiring specific roles
func Authorize(roles ...models.UserRole) *AuthorizationDecorator {
	return &AuthorizationDecorator{
		Roles:       roles,
		Description: "Requires specific roles",
	}
}

// AuthorizeAdmin decorator for requiring admin privileges
func AuthorizeAdmin() *AuthorizationDecorator {
	return &AuthorizationDecorator{
		Roles:       []models.UserRole{models.RoleAdministrator, models.RoleManager},
		Description: "Requires admin privileges",
	}
}

// AuthorizeAgent decorator for requiring agent privileges
func AuthorizeAgent() *AuthorizationDecorator {
	return &AuthorizationDecorator{
		Roles:       []models.UserRole{models.RoleSupportAgent, models.RoleAdministrator, models.RoleManager},
		Description: "Requires agent privileges",
	}
}

// AuthorizePermission decorator for requiring specific permissions
func AuthorizePermission(permissions ...string) *AuthorizationDecorator {
	return &AuthorizationDecorator{
		Permissions: permissions,
		Description: "Requires specific permissions",
	}
}

// AuthorizeAny decorator for requiring any of the specified roles or permissions
func AuthorizeAny(roles []models.UserRole, permissions []string) *AuthorizationDecorator {
	return &AuthorizationDecorator{
		Roles:       roles,
		Permissions: permissions,
		Description: "Requires any of the specified roles or permissions",
	}
}

// DecoratorMiddleware creates middleware from authorization decorators
func (m *AuthMiddleware) DecoratorMiddleware(decorator *AuthorizationDecorator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(401, "user not found in context")
			}

			// Check roles if specified
			if len(decorator.Roles) > 0 {
				hasRole := false
				for _, role := range decorator.Roles {
					if user.Role == role {
						hasRole = true
						break
					}
				}
				if hasRole {
					return next(c)
				}
			}

			// Check permissions if specified
			if len(decorator.Permissions) > 0 {
				for _, permission := range decorator.Permissions {
					if m.HasPermission(user.Role, permission) {
						return next(c)
					}
				}
			}

			// If no roles or permissions specified, deny access
			if len(decorator.Roles) == 0 && len(decorator.Permissions) == 0 {
				return echo.NewHTTPError(403, "insufficient permissions")
			}

			return echo.NewHTTPError(403, "insufficient permissions")
		}
	}
}

// MultiDecoratorMiddleware creates middleware from multiple authorization decorators
// All decorators must pass for the request to proceed (AND logic)
func (m *AuthMiddleware) MultiDecoratorMiddleware(decorators ...*AuthorizationDecorator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(401, "user not found in context")
			}

			// Check all decorators - ALL must pass
			for _, decorator := range decorators {
				if decorator == nil {
					continue
				}

				hasAccess := false

				// Check roles if specified
				if len(decorator.Roles) > 0 {
					for _, role := range decorator.Roles {
						if user.Role == role {
							hasAccess = true
							break
						}
					}
				}

				// Check permissions if specified
				if !hasAccess && len(decorator.Permissions) > 0 {
					for _, permission := range decorator.Permissions {
						if m.HasPermission(user.Role, permission) {
							hasAccess = true
							break
						}
					}
				}

				// If no roles or permissions specified, deny access
				if len(decorator.Roles) == 0 && len(decorator.Permissions) == 0 {
					return echo.NewHTTPError(403, "insufficient permissions")
				}

				// If this decorator doesn't grant access, deny the request
				if !hasAccess {
					return echo.NewHTTPError(403, "insufficient permissions")
				}
			}

			return next(c)
		}
	}
}

// HandlerDecorator wraps a handler function with authorization decorators
type HandlerDecorator struct {
	Handler     echo.HandlerFunc
	Decorators  []*AuthorizationDecorator
	Description string
}

// NewHandlerDecorator creates a new handler decorator with a single decorator
func NewHandlerDecorator(handler echo.HandlerFunc, decorator *AuthorizationDecorator) *HandlerDecorator {
	return &HandlerDecorator{
		Handler:     handler,
		Decorators:  []*AuthorizationDecorator{decorator},
		Description: decorator.Description,
	}
}

// NewMultiHandlerDecorator creates a new handler decorator with multiple decorators
func NewMultiHandlerDecorator(handler echo.HandlerFunc, decorators ...*AuthorizationDecorator) *HandlerDecorator {
	descriptions := make([]string, 0)
	for _, decorator := range decorators {
		if decorator != nil {
			descriptions = append(descriptions, decorator.Description)
		}
	}

	return &HandlerDecorator{
		Handler:     handler,
		Decorators:  decorators,
		Description: strings.Join(descriptions, " AND "),
	}
}

// AddDecorator adds an additional decorator to the handler
func (hd *HandlerDecorator) AddDecorator(decorator *AuthorizationDecorator) *HandlerDecorator {
	hd.Decorators = append(hd.Decorators, decorator)
	if decorator != nil {
		if hd.Description == "" {
			hd.Description = decorator.Description
		} else {
			hd.Description += " AND " + decorator.Description
		}
	}
	return hd
}

// ToHandlerFunc converts the decorator to an Echo handler function
func (hd *HandlerDecorator) ToHandlerFunc(authMiddleware *AuthMiddleware) echo.HandlerFunc {
	if len(hd.Decorators) == 0 {
		return hd.Handler
	}

	// Filter out nil decorators
	validDecorators := make([]*AuthorizationDecorator, 0)
	for _, decorator := range hd.Decorators {
		if decorator != nil {
			validDecorators = append(validDecorators, decorator)
		}
	}

	if len(validDecorators) == 0 {
		return hd.Handler
	}

	if len(validDecorators) == 1 {
		middleware := authMiddleware.DecoratorMiddleware(validDecorators[0])
		return middleware(hd.Handler)
	}

	// Multiple decorators - use multi-decorator middleware
	middleware := authMiddleware.MultiDecoratorMiddleware(validDecorators...)
	return middleware(hd.Handler)
}

// RouteDecorator represents a route with authorization decorators
type RouteDecorator struct {
	Method      string
	Path        string
	Handler     *HandlerDecorator
	Description string
}

// NewRouteDecorator creates a new route decorator
func NewRouteDecorator(method, path string, handler *HandlerDecorator) *RouteDecorator {
	return &RouteDecorator{
		Method:      method,
		Path:        path,
		Handler:     handler,
		Description: handler.Description,
	}
}

// DecoratedRouteGroup manages routes with authorization decorators
type DecoratedRouteGroup struct {
	group          *echo.Group
	authMiddleware *AuthMiddleware
	routes         []*RouteDecorator
}

// NewDecoratedRouteGroup creates a new decorated route group
func NewDecoratedRouteGroup(group *echo.Group, authMiddleware *AuthMiddleware) *DecoratedRouteGroup {
	return &DecoratedRouteGroup{
		group:          group,
		authMiddleware: authMiddleware,
		routes:         make([]*RouteDecorator, 0),
	}
}

// AddRoute adds a route with authorization decorator
func (drg *DecoratedRouteGroup) AddRoute(method, path string, handler *HandlerDecorator) {
	route := NewRouteDecorator(method, path, handler)
	drg.routes = append(drg.routes, route)

	// Register the route with Echo
	drg.group.Add(method, path, handler.ToHandlerFunc(drg.authMiddleware))
}

// GET adds a GET route with authorization
func (drg *DecoratedRouteGroup) GET(path string, handler *HandlerDecorator) {
	drg.AddRoute("GET", path, handler)
}

// POST adds a POST route with authorization
func (drg *DecoratedRouteGroup) POST(path string, handler *HandlerDecorator) {
	drg.AddRoute("POST", path, handler)
}

// PUT adds a PUT route with authorization
func (drg *DecoratedRouteGroup) PUT(path string, handler *HandlerDecorator) {
	drg.AddRoute("PUT", path, handler)
}

// DELETE adds a DELETE route with authorization
func (drg *DecoratedRouteGroup) DELETE(path string, handler *HandlerDecorator) {
	drg.AddRoute("DELETE", path, handler)
}

// PATCH adds a PATCH route with authorization
func (drg *DecoratedRouteGroup) PATCH(path string, handler *HandlerDecorator) {
	drg.AddRoute("PATCH", path, handler)
}

// GetRoutes returns all registered routes
func (drg *DecoratedRouteGroup) GetRoutes() []*RouteDecorator {
	return drg.routes
}

// GetGroup returns the underlying Echo group
func (drg *DecoratedRouteGroup) GetGroup() *echo.Group {
	return drg.group
}
