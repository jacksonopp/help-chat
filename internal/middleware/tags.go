package middleware

import (
	"reflect"
	"strings"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"github.com/labstack/echo/v4"
)

// AuthorizeTag represents authorization requirements in struct tags
type AuthorizeTag struct {
	Roles       []models.UserRole
	Permissions []string
	Policy      string
}

// ParseAuthorizeTag parses authorization requirements from struct tags
func ParseAuthorizeTag(tag string) *AuthorizeTag {
	if tag == "" {
		return nil
	}

	parts := strings.Split(tag, ",")
	authorize := &AuthorizeTag{}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if it's a role
		if role := parseRole(part); role != "" {
			authorize.Roles = append(authorize.Roles, role)
			continue
		}

		// Check if it's a permission
		if strings.Contains(part, ":") {
			authorize.Permissions = append(authorize.Permissions, part)
			continue
		}

		// Check if it's a policy
		if strings.HasPrefix(part, "Policy=") {
			authorize.Policy = strings.TrimPrefix(part, "Policy=")
			continue
		}
	}

	return authorize
}

// ParseMultipleAuthorizeTags parses multiple authorization tags (AND logic)
func ParseMultipleAuthorizeTags(tags ...string) []*AuthorizeTag {
	var authorizeTags []*AuthorizeTag
	for _, tag := range tags {
		if parsed := ParseAuthorizeTag(tag); parsed != nil {
			authorizeTags = append(authorizeTags, parsed)
		}
	}
	return authorizeTags
}

// parseRole converts a string to UserRole
func parseRole(roleStr string) models.UserRole {
	switch strings.ToUpper(roleStr) {
	case "ADMIN", "ADMINISTRATOR":
		return models.RoleAdministrator
	case "MANAGER":
		return models.RoleManager
	case "AGENT", "SUPPORT_AGENT":
		return models.RoleSupportAgent
	case "USER", "END_USER":
		return models.RoleEndUser
	default:
		return ""
	}
}

// TaggedHandler represents a handler with authorization tags
type TaggedHandler struct {
	Handler echo.HandlerFunc
	Tags    []*AuthorizeTag
}

// NewTaggedHandler creates a new tagged handler with a single tag
func NewTaggedHandler(handler echo.HandlerFunc, tag *AuthorizeTag) *TaggedHandler {
	return &TaggedHandler{
		Handler: handler,
		Tags:    []*AuthorizeTag{tag},
	}
}

// NewMultiTaggedHandler creates a new tagged handler with multiple tags
func NewMultiTaggedHandler(handler echo.HandlerFunc, tags ...*AuthorizeTag) *TaggedHandler {
	return &TaggedHandler{
		Handler: handler,
		Tags:    tags,
	}
}

// AddTag adds an additional tag to the handler
func (th *TaggedHandler) AddTag(tag *AuthorizeTag) *TaggedHandler {
	th.Tags = append(th.Tags, tag)
	return th
}

// ToHandlerFunc converts the tagged handler to an Echo handler function
func (th *TaggedHandler) ToHandlerFunc(authMiddleware *AuthMiddleware) echo.HandlerFunc {
	if len(th.Tags) == 0 {
		return th.Handler
	}

	// Filter out nil tags
	validTags := make([]*AuthorizeTag, 0)
	for _, tag := range th.Tags {
		if tag != nil {
			validTags = append(validTags, tag)
		}
	}

	if len(validTags) == 0 {
		return th.Handler
	}

	if len(validTags) == 1 {
		return func(c echo.Context) error {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(401, "user not found in context")
			}

			tag := validTags[0]

			// Check roles if specified
			if len(tag.Roles) > 0 {
				hasRole := false
				for _, role := range tag.Roles {
					if user.Role == role {
						hasRole = true
						break
					}
				}
				if hasRole {
					return th.Handler(c)
				}
			}

			// Check permissions if specified
			if len(tag.Permissions) > 0 {
				for _, permission := range tag.Permissions {
					if authMiddleware.HasPermission(user.Role, permission) {
						return th.Handler(c)
					}
				}
			}

			// Check policy if specified
			if tag.Policy != "" {
				if authMiddleware.HasPermission(user.Role, tag.Policy) {
					return th.Handler(c)
				}
			}

			// If no authorization requirements specified, deny access
			if len(tag.Roles) == 0 && len(tag.Permissions) == 0 && tag.Policy == "" {
				return echo.NewHTTPError(403, "insufficient permissions")
			}

			return echo.NewHTTPError(403, "insufficient permissions")
		}
	}

	// Multiple tags - ALL must pass (AND logic)
	return func(c echo.Context) error {
		user := c.Get("user").(*models.User)
		if user == nil {
			return echo.NewHTTPError(401, "user not found in context")
		}

		// Check all tags - ALL must pass
		for _, tag := range validTags {
			hasAccess := false

			// Check roles if specified
			if len(tag.Roles) > 0 {
				for _, role := range tag.Roles {
					if user.Role == role {
						hasAccess = true
						break
					}
				}
			}

			// Check permissions if specified
			if !hasAccess && len(tag.Permissions) > 0 {
				for _, permission := range tag.Permissions {
					if authMiddleware.HasPermission(user.Role, permission) {
						hasAccess = true
						break
					}
				}
			}

			// Check policy if specified
			if !hasAccess && tag.Policy != "" {
				if authMiddleware.HasPermission(user.Role, tag.Policy) {
					hasAccess = true
				}
			}

			// If no authorization requirements specified, deny access
			if len(tag.Roles) == 0 && len(tag.Permissions) == 0 && tag.Policy == "" {
				return echo.NewHTTPError(403, "insufficient permissions")
			}

			// If this tag doesn't grant access, deny the request
			if !hasAccess {
				return echo.NewHTTPError(403, "insufficient permissions")
			}
		}

		return th.Handler(c)
	}
}

// TaggedRouteGroup manages routes with authorization tags
type TaggedRouteGroup struct {
	group          *echo.Group
	authMiddleware *AuthMiddleware
	routes         []*TaggedRoute
}

// TaggedRoute represents a route with authorization tags
type TaggedRoute struct {
	Method  string
	Path    string
	Handler *TaggedHandler
	Tags    []string
}

// NewTaggedRouteGroup creates a new tagged route group
func NewTaggedRouteGroup(group *echo.Group, authMiddleware *AuthMiddleware) *TaggedRouteGroup {
	return &TaggedRouteGroup{
		group:          group,
		authMiddleware: authMiddleware,
		routes:         make([]*TaggedRoute, 0),
	}
}

// AddRoute adds a route with authorization tag
func (trg *TaggedRouteGroup) AddRoute(method, path string, handler *TaggedHandler) {
	route := &TaggedRoute{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
	trg.routes = append(trg.routes, route)

	// Register the route with Echo
	trg.group.Add(method, path, handler.ToHandlerFunc(trg.authMiddleware))
}

// GET adds a GET route with authorization
func (trg *TaggedRouteGroup) GET(path string, handler *TaggedHandler) {
	trg.AddRoute("GET", path, handler)
}

// POST adds a POST route with authorization
func (trg *TaggedRouteGroup) POST(path string, handler *TaggedHandler) {
	trg.AddRoute("POST", path, handler)
}

// PUT adds a PUT route with authorization
func (trg *TaggedRouteGroup) PUT(path string, handler *TaggedHandler) {
	trg.AddRoute("PUT", path, handler)
}

// DELETE adds a DELETE route with authorization
func (trg *TaggedRouteGroup) DELETE(path string, handler *TaggedHandler) {
	trg.AddRoute("DELETE", path, handler)
}

// PATCH adds a PATCH route with authorization
func (trg *TaggedRouteGroup) PATCH(path string, handler *TaggedHandler) {
	trg.AddRoute("PATCH", path, handler)
}

// GetRoutes returns all registered routes
func (trg *TaggedRouteGroup) GetRoutes() []*TaggedRoute {
	return trg.routes
}

// GetGroup returns the underlying Echo group
func (trg *TaggedRouteGroup) GetGroup() *echo.Group {
	return trg.group
}

// TaggedHandlerStruct represents a handler struct with authorization tags
type TaggedHandlerStruct struct {
	Handler interface{}
	Tags    []string
}

// NewTaggedHandlerStruct creates a new tagged handler struct with a single tag
func NewTaggedHandlerStruct(handler interface{}, tag string) *TaggedHandlerStruct {
	return &TaggedHandlerStruct{
		Handler: handler,
		Tags:    []string{tag},
	}
}

// NewMultiTaggedHandlerStruct creates a new tagged handler struct with multiple tags
func NewMultiTaggedHandlerStruct(handler interface{}, tags ...string) *TaggedHandlerStruct {
	return &TaggedHandlerStruct{
		Handler: handler,
		Tags:    tags,
	}
}

// AddTag adds an additional tag to the handler struct
func (ths *TaggedHandlerStruct) AddTag(tag string) *TaggedHandlerStruct {
	ths.Tags = append(ths.Tags, tag)
	return ths
}

// ToHandlerFunc converts the tagged handler struct to an Echo handler function
func (ths *TaggedHandlerStruct) ToHandlerFunc(authMiddleware *AuthMiddleware) echo.HandlerFunc {
	// Parse all authorization tags
	authorizeTags := ParseMultipleAuthorizeTags(ths.Tags...)

	// Get the handler function using reflection
	handlerValue := reflect.ValueOf(ths.Handler)
	if handlerValue.Kind() != reflect.Func {
		panic("handler must be a function")
	}

	handlerFunc := func(c echo.Context) error {
		// Check authorization if tags are specified
		if len(authorizeTags) > 0 {
			user := c.Get("user").(*models.User)
			if user == nil {
				return echo.NewHTTPError(401, "user not found in context")
			}

			// Check all tags - ALL must pass (AND logic)
			for _, authorizeTag := range authorizeTags {
				hasAccess := false

				// Check roles if specified
				if len(authorizeTag.Roles) > 0 {
					for _, role := range authorizeTag.Roles {
						if user.Role == role {
							hasAccess = true
							break
						}
					}
				}

				// Check permissions if specified
				if !hasAccess && len(authorizeTag.Permissions) > 0 {
					for _, permission := range authorizeTag.Permissions {
						if authMiddleware.HasPermission(user.Role, permission) {
							hasAccess = true
							break
						}
					}
				}

				// Check policy if specified
				if !hasAccess && authorizeTag.Policy != "" {
					if authMiddleware.HasPermission(user.Role, authorizeTag.Policy) {
						hasAccess = true
					}
				}

				// If no authorization requirements specified, deny access
				if len(authorizeTag.Roles) == 0 && len(authorizeTag.Permissions) == 0 && authorizeTag.Policy == "" {
					return echo.NewHTTPError(403, "insufficient permissions")
				}

				// If this tag doesn't grant access, deny the request
				if !hasAccess {
					return echo.NewHTTPError(403, "insufficient permissions")
				}
			}
		}

		// Call the handler function
		args := []reflect.Value{reflect.ValueOf(c)}
		results := handlerValue.Call(args)

		// Handle the return value
		if len(results) > 0 {
			if err := results[0].Interface(); err != nil {
				return err.(error)
			}
		}

		return nil
	}

	return handlerFunc
}
