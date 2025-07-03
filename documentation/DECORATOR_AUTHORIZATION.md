# Decorator-Style Authorization

This document describes the decorator-style authorization system that provides a .NET-like experience for adding authorization to routes without modifying the `AuthorizationConfig`.

## Overview

The decorator-style authorization system provides two approaches:

1. **Function Decorators**: Similar to .NET's `[Authorize]` attribute
2. **Tag-Based Authorization**: Using string tags for authorization requirements

Both approaches allow you to add authorization directly to handler methods without centralized configuration.

## Function Decorators

### Basic Usage

```go
// RegisterDecoratedRoutes demonstrates decorator-style authorization
func (h *DecoratedTicketHandler) RegisterDecoratedRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
    tickets := e.Group("/api/v1/decorated-tickets")
    tickets.Use(authMiddlewareInstance.Authenticate)

    decoratedGroup := authMiddleware.NewDecoratedRouteGroup(tickets, authMiddlewareInstance)

    // Admin-only route
    decoratedGroup.DELETE("/:id", authMiddleware.NewHandlerDecorator(
        h.DeleteTicketDecorated,
        authMiddleware.AuthorizeAdmin(),
    ))

    // Agent-only route
    decoratedGroup.POST("/:id/assign", authMiddleware.NewHandlerDecorator(
        h.AssignTicketDecorated,
        authMiddleware.AuthorizeAgent(),
    ))
}
```

### Available Decorators

#### Role-Based Decorators

```go
// Require specific roles
authMiddleware.Authorize(models.RoleAdministrator, models.RoleManager)

// Require admin privileges (ADMIN or MANAGER)
authMiddleware.AuthorizeAdmin()

// Require agent privileges (AGENT, ADMIN, or MANAGER)
authMiddleware.AuthorizeAgent()
```

#### Permission-Based Decorators

```go
// Require specific permissions
authMiddleware.AuthorizePermission("ticket:delete")

// Require any of multiple permissions
authMiddleware.AuthorizePermission("ticket:escalate", "ticket:admin")
```

#### Mixed Decorators

```go
// Require any of the specified roles OR permissions
authMiddleware.AuthorizeAny(
    []models.UserRole{models.RoleAdministrator, models.RoleManager},
    []string{"ticket:update", "ticket:admin"},
)
```

### Complete Example

```go
func (h *DecoratedTicketHandler) RegisterDecoratedRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
    tickets := e.Group("/api/v1/decorated-tickets")
    tickets.Use(authMiddlewareInstance.Authenticate)

    decoratedGroup := authMiddleware.NewDecoratedRouteGroup(tickets, authMiddlewareInstance)

    // Example 1: Admin-only route
    decoratedGroup.DELETE("/:id", authMiddleware.NewHandlerDecorator(
        h.DeleteTicketDecorated,
        authMiddleware.AuthorizeAdmin(),
    ))

    // Example 2: Agent-only route
    decoratedGroup.POST("/:id/assign", authMiddleware.NewHandlerDecorator(
        h.AssignTicketDecorated,
        authMiddleware.AuthorizeAgent(),
    ))

    // Example 3: Specific roles
    decoratedGroup.POST("/:id/status", authMiddleware.NewHandlerDecorator(
        h.UpdateTicketStatusDecorated,
        authMiddleware.Authorize(models.RoleSupportAgent, models.RoleManager, models.RoleAdministrator),
    ))

    // Example 4: Permission-based
    decoratedGroup.GET("/stats", authMiddleware.NewHandlerDecorator(
        h.GetTicketStatsDecorated,
        authMiddleware.AuthorizePermission("ticket:stats:read"),
    ))

    // Example 5: Multiple permissions
    decoratedGroup.POST("/:id/escalate", authMiddleware.NewHandlerDecorator(
        h.EscalateTicketDecorated,
        authMiddleware.AuthorizePermission("ticket:escalate", "ticket:admin"),
    ))

    // Example 6: Mixed roles and permissions
    decoratedGroup.PUT("/:id", authMiddleware.NewHandlerDecorator(
        h.UpdateTicketDecorated,
        authMiddleware.AuthorizeAny(
            []models.UserRole{models.RoleAdministrator, models.RoleManager},
            []string{"ticket:update", "ticket:admin"},
        ),
    ))

    // Example 7: No additional authorization required
    decoratedGroup.GET("/:id", authMiddleware.NewHandlerDecorator(
        h.GetTicketDecorated,
        nil, // No decorator means no additional authorization required
    ))
}
```

## Tag-Based Authorization

### Basic Usage

Tag-based authorization uses string tags similar to .NET's `[Authorize("ADMIN")]`:

```go
func (h *TaggedTicketHandler) RegisterTaggedRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
    tickets := e.Group("/api/v1/tagged-tickets")
    tickets.Use(authMiddlewareInstance.Authenticate)

    taggedGroup := authMiddleware.NewTaggedRouteGroup(tickets, authMiddlewareInstance)

    // Admin-only route using role tag
    taggedGroup.DELETE("/:id", authMiddleware.NewTaggedHandler(
        h.DeleteTicketTagged,
        authMiddleware.ParseAuthorizeTag("ADMIN,MANAGER"),
    ))

    // Permission-based authorization
    taggedGroup.GET("/stats", authMiddleware.NewTaggedHandler(
        h.GetTicketStatsTagged,
        authMiddleware.ParseAuthorizeTag("ticket:stats:read"),
    ))
}
```

### Tag Format

Tags support multiple formats:

#### Role Tags

```go
// Single role
"ADMIN"

// Multiple roles (any of them)
"ADMIN,MANAGER"

// Support agent or higher
"AGENT,ADMIN,MANAGER"
```

#### Permission Tags

```go
// Single permission
"ticket:delete"

// Multiple permissions (any of them)
"ticket:escalate,ticket:admin"
```

#### Mixed Tags

```go
// Roles and permissions (any of them)
"ADMIN,MANAGER,ticket:update"
```

#### Policy Tags

```go
// Policy-based authorization
"Policy=ticket:status:update"
```

### Complete Tag Examples

```go
func (h *TaggedTicketHandler) RegisterTaggedRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
    tickets := e.Group("/api/v1/tagged-tickets")
    tickets.Use(authMiddlewareInstance.Authenticate)

    taggedGroup := authMiddleware.NewTaggedRouteGroup(tickets, authMiddlewareInstance)

    // Example 1: Admin-only route using role tag
    taggedGroup.DELETE("/:id", authMiddleware.NewTaggedHandler(
        h.DeleteTicketTagged,
        authMiddleware.ParseAuthorizeTag("ADMIN,MANAGER"),
    ))

    // Example 2: Agent-only route using role tag
    taggedGroup.POST("/:id/assign", authMiddleware.NewTaggedHandler(
        h.AssignTicketTagged,
        authMiddleware.ParseAuthorizeTag("AGENT,ADMIN,MANAGER"),
    ))

    // Example 3: Permission-based authorization
    taggedGroup.GET("/stats", authMiddleware.NewTaggedHandler(
        h.GetTicketStatsTagged,
        authMiddleware.ParseAuthorizeTag("ticket:stats:read"),
    ))

    // Example 4: Multiple permissions (any of them)
    taggedGroup.POST("/:id/escalate", authMiddleware.NewTaggedHandler(
        h.EscalateTicketTagged,
        authMiddleware.ParseAuthorizeTag("ticket:escalate,ticket:admin"),
    ))

    // Example 5: Mixed roles and permissions
    taggedGroup.PUT("/:id", authMiddleware.NewTaggedHandler(
        h.UpdateTicketTagged,
        authMiddleware.ParseAuthorizeTag("ADMIN,MANAGER,ticket:update"),
    ))

    // Example 6: Policy-based authorization
    taggedGroup.POST("/:id/status", authMiddleware.NewTaggedHandler(
        h.UpdateTicketStatusTagged,
        authMiddleware.ParseAuthorizeTag("Policy=ticket:status:update"),
    ))

    // Example 7: No authorization required
    taggedGroup.GET("/:id", authMiddleware.NewTaggedHandler(
        h.GetTicketTagged,
        nil, // No tag means no additional authorization required
    ))

    // Example 8: Own tickets permission
    taggedGroup.GET("/my", authMiddleware.NewTaggedHandler(
        h.GetMyTicketsTagged,
        authMiddleware.ParseAuthorizeTag("ticket:read:own"),
    ))

    // Example 9: Complex authorization with multiple requirements
    taggedGroup.POST("/:id/comment", authMiddleware.NewTaggedHandler(
        h.AddCommentTagged,
        authMiddleware.ParseAuthorizeTag("AGENT,ADMIN,MANAGER,ticket:comment:add"),
    ))
}
```

## Comparison with .NET

### .NET ASP.NET Core

```csharp
[Authorize(Roles = "Admin")]
[HttpDelete("{id}")]
public IActionResult DeleteTicket(int id) { ... }

[Authorize(Policy = "TicketDelete")]
[HttpDelete("{id}")]
public IActionResult DeleteTicket(int id) { ... }

[Authorize(Roles = "Admin,Manager")]
[Authorize(Policy = "TicketUpdate")]
[HttpPut("{id}")]
public IActionResult UpdateTicket(int id) { ... }
```

### Go with Decorators

```go
// Function decorators
decoratedGroup.DELETE("/:id", authMiddleware.NewHandlerDecorator(
    h.DeleteTicketDecorated,
    authMiddleware.AuthorizeAdmin(),
))

decoratedGroup.DELETE("/:id", authMiddleware.NewHandlerDecorator(
    h.DeleteTicketDecorated,
    authMiddleware.AuthorizePermission("ticket:delete"),
))

decoratedGroup.PUT("/:id", authMiddleware.NewHandlerDecorator(
    h.UpdateTicketDecorated,
    authMiddleware.AuthorizeAny(
        []models.UserRole{models.RoleAdministrator, models.RoleManager},
        []string{"ticket:update"},
    ),
))
```

### Go with Tags

```go
// Tag-based (most similar to .NET)
taggedGroup.DELETE("/:id", authMiddleware.NewTaggedHandler(
    h.DeleteTicketTagged,
    authMiddleware.ParseAuthorizeTag("ADMIN"),
))

taggedGroup.DELETE("/:id", authMiddleware.NewTaggedHandler(
    h.DeleteTicketTagged,
    authMiddleware.ParseAuthorizeTag("Policy=ticket:delete"),
))

taggedGroup.PUT("/:id", authMiddleware.NewTaggedHandler(
    h.UpdateTicketTagged,
    authMiddleware.ParseAuthorizeTag("ADMIN,MANAGER,ticket:update"),
))
```

## Benefits

### 1. **Declarative Authorization**

- Authorization requirements are clearly visible at the route level
- No need to modify centralized configuration
- Easy to understand and maintain

### 2. **Flexible and Expressive**

- Support for roles, permissions, and policies
- Multiple authorization strategies in one place
- Easy to combine different authorization requirements

### 3. **Developer-Friendly**

- Familiar syntax for .NET developers
- Intuitive and readable
- IDE-friendly with autocomplete support

### 4. **Maintainable**

- Authorization logic is co-located with handlers
- Easy to modify individual route requirements
- Clear separation of concerns

## Best Practices

### 1. **Choose the Right Approach**

- Use **function decorators** for complex authorization logic
- Use **tag-based** for simple, .NET-like syntax
- Use **traditional middleware** for global authorization patterns

### 2. **Be Specific**

```go
// Good: Specific permission
authMiddleware.AuthorizePermission("ticket:delete")

// Avoid: Too broad
authMiddleware.AuthorizeAdmin()
```

### 3. **Use Meaningful Permissions**

```go
// Good: Descriptive permissions
"ticket:delete:own"
"ticket:update:assigned"

// Avoid: Generic permissions
"write"
"read"
```

### 4. **Document Authorization Requirements**

```go
// Add comments to explain complex authorization
// Requires admin OR manager with ticket:update permission
taggedGroup.PUT("/:id", authMiddleware.NewTaggedHandler(
    h.UpdateTicketTagged,
    authMiddleware.ParseAuthorizeTag("ADMIN,MANAGER,ticket:update"),
))
```

## Migration Guide

### From Traditional Middleware

**Before:**

```go
func (h *TicketHandler) RegisterRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
    tickets := e.Group("/api/v1/tickets")
    tickets.Use(authMiddlewareInstance.Authenticate)

    tickets.DELETE("/:id", h.DeleteTicket, authMiddlewareInstance.RequireAdmin())
    tickets.POST("/:id/assign", h.AssignTicket, authMiddlewareInstance.RequireAgent())
}
```

**After (Function Decorators):**

```go
func (h *DecoratedTicketHandler) RegisterDecoratedRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
    tickets := e.Group("/api/v1/decorated-tickets")
    tickets.Use(authMiddlewareInstance.Authenticate)

    decoratedGroup := authMiddleware.NewDecoratedRouteGroup(tickets, authMiddlewareInstance)

    decoratedGroup.DELETE("/:id", authMiddleware.NewHandlerDecorator(
        h.DeleteTicketDecorated,
        authMiddleware.AuthorizeAdmin(),
    ))
    decoratedGroup.POST("/:id/assign", authMiddleware.NewHandlerDecorator(
        h.AssignTicketDecorated,
        authMiddleware.AuthorizeAgent(),
    ))
}
```

**After (Tag-Based):**

```go
func (h *TaggedTicketHandler) RegisterTaggedRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
    tickets := e.Group("/api/v1/tagged-tickets")
    tickets.Use(authMiddlewareInstance.Authenticate)

    taggedGroup := authMiddleware.NewTaggedRouteGroup(tickets, authMiddlewareInstance)

    taggedGroup.DELETE("/:id", authMiddleware.NewTaggedHandler(
        h.DeleteTicketTagged,
        authMiddleware.ParseAuthorizeTag("ADMIN"),
    ))
    taggedGroup.POST("/:id/assign", authMiddleware.NewTaggedHandler(
        h.AssignTicketTagged,
        authMiddleware.ParseAuthorizeTag("AGENT,ADMIN,MANAGER"),
    ))
}
```

## Testing

### Testing Decorator Authorization

```go
func TestDecoratorAuthorization(t *testing.T) {
    // Setup test environment
    authMiddlewareInstance := authMiddleware.NewAuthMiddleware(authService)

    // Test admin-only route
    t.Run("Admin can access admin-only route", func(t *testing.T) {
        // Test implementation
    })

    // Test agent-only route
    t.Run("Agent can access agent-only route", func(t *testing.T) {
        // Test implementation
    })

    // Test permission-based route
    t.Run("User with permission can access permission-based route", func(t *testing.T) {
        // Test implementation
    })
}
```

## Conclusion

The decorator-style authorization system provides a powerful and flexible way to add authorization to your routes. It offers:

- **Familiar syntax** for .NET developers
- **Declarative approach** to authorization
- **Flexible combinations** of roles and permissions
- **Easy maintenance** and modification
- **Clear visibility** of authorization requirements

Choose the approach that best fits your team's preferences and project requirements. Both function decorators and tag-based authorization provide excellent alternatives to traditional middleware-based authorization.
