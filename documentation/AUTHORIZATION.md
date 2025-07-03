# Authorization System

This document describes the authorization system implemented in the HelpChat application.

## Overview

The authorization system provides role-based access control (RBAC) with granular permissions for different API endpoints. It consists of multiple layers of security:

1. **Middleware Layer**: Route-level authorization using Echo middleware
2. **Service Layer**: Business logic authorization checks
3. **Permission System**: Granular permission-based access control

## User Roles

The system defines four user roles with different levels of access:

### 1. END_USER

- **Description**: Regular users who can create and manage their own tickets
- **Permissions**:
  - `ticket:create` - Create new tickets
  - `ticket:read:own` - View their own tickets
  - `ticket:update:own` - Update their own tickets

### 2. SUPPORT_AGENT

- **Description**: Support staff who can handle tickets
- **Permissions**:
  - All END_USER permissions
  - `ticket:read` - View all tickets
  - `ticket:update` - Update any ticket
  - `ticket:assign` - Assign tickets to agents
  - `ticket:status:update` - Update ticket status
  - `ticket:escalate` - Escalate tickets
  - `ticket:stats:read` - View ticket statistics

### 3. MANAGER

- **Description**: Team managers with administrative capabilities
- **Permissions**:
  - All SUPPORT_AGENT permissions
  - `ticket:delete` - Delete tickets
  - `user:manage` - Manage user accounts

### 4. ADMINISTRATOR

- **Description**: System administrators with full access
- **Permissions**:
  - All MANAGER permissions
  - `system:admin` - Full system access

## Implementation

### 1. Route-Level Authorization

Routes are secured using middleware in the `RegisterRoutes` method:

```go
func (h *TicketHandler) RegisterRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
    tickets := e.Group("/api/v1/tickets")
    tickets.Use(authMiddlewareInstance.Authenticate)

    // Public routes (require authentication only)
    tickets.GET("", h.ListTickets)
    tickets.POST("", h.CreateTicket)
    tickets.GET("/:id", h.GetTicket)
    tickets.PUT("/:id", h.UpdateTicket)

    // Admin-only routes
    tickets.DELETE("/:id", h.DeleteTicket, authMiddlewareInstance.RequireAdmin())

    // Agent-only routes
    tickets.POST("/:id/assign", h.AssignTicket, authMiddlewareInstance.RequireAgent())
    tickets.POST("/:id/status", h.UpdateTicketStatus, authMiddlewareInstance.RequireAgent())
    tickets.POST("/:id/escalate", h.EscalateTicket, authMiddlewareInstance.RequireAgent())
    tickets.GET("/stats", h.GetTicketStats, authMiddlewareInstance.RequireAgent())
}
```

### 2. Available Middleware Functions

#### Basic Role Middleware

- `RequireRole(role)` - Requires a specific role
- `RequireAnyRole(roles...)` - Requires any of the specified roles
- `RequireAdmin()` - Requires administrator or manager role
- `RequireAgent()` - Requires agent, manager, or administrator role
- `RequireManager()` - Requires manager or administrator role

#### Advanced Middleware

- `RequireOwnerOrAdmin(ownerIDGetter)` - Allows access if user owns the resource or is admin
- `RequirePermission(permission)` - Checks for specific permissions

### 3. Service Layer Authorization

As a backup security measure, the service layer also includes authorization checks:

```go
func (s *TicketService) DeleteTicket(ctx context.Context, ticketID uuid.UUID, userID uuid.UUID) error {
    // Get user to check authorization
    user, err := s.userRepo.GetByID(userID.String())
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }

    // Only admins can delete tickets
    if !user.IsAdmin() {
        return fmt.Errorf("insufficient permissions: only administrators can delete tickets")
    }

    // ... rest of the logic
}
```

## Permission System

The permission system provides granular control over what actions users can perform:

### Permission Format

Permissions follow the format: `resource:action[:scope]`

Examples:

- `ticket:create` - Create tickets
- `ticket:read:own` - Read own tickets only
- `ticket:delete` - Delete tickets
- `user:manage` - Manage user accounts

### Adding New Permissions

To add new permissions:

1. **Update the permission map** in `internal/middleware/auth.go`:

```go
permissions := map[models.UserRole][]string{
    models.RoleEndUser: {
        "ticket:create",
        "ticket:read:own",
        "ticket:update:own",
        "new:permission", // Add new permission here
    },
    // ... other roles
}
```

2. **Add middleware to routes** that require the permission:

```go
tickets.POST("/new-endpoint", h.NewHandler, authMiddlewareInstance.RequirePermission("new:permission"))
```

## Security Best Practices

### 1. Defense in Depth

- Use both middleware and service-layer authorization
- Validate permissions at multiple levels

### 2. Principle of Least Privilege

- Grant users only the permissions they need
- Use the most restrictive permission possible

### 3. Regular Auditing

- Review permissions regularly
- Monitor access patterns
- Log authorization failures

### 4. Secure Defaults

- Default to denying access
- Explicitly grant permissions
- Validate all user inputs

## Testing Authorization

The authorization system includes comprehensive tests in `test/authorization_test.go`:

```bash
go test ./test -v -run TestAuthorizationMiddleware
go test ./test -v -run TestPermissionSystem
```

## Configuration

Authorization can be configured through the `AuthorizationConfig` in `internal/middleware/authorization.go`:

```go
config := authMiddleware.NewAuthorizationConfig()
// Add custom route permissions
config.RoutePermissions = append(config.RoutePermissions, authMiddleware.RoutePermission{
    Method:      "POST",
    Path:        "/api/v1/custom",
    Permission:  "custom:action",
    Description: "Custom action",
})
```

## Error Handling

Authorization failures return appropriate HTTP status codes:

- **401 Unauthorized**: User not authenticated
- **403 Forbidden**: User authenticated but lacks permission
- **400 Bad Request**: Invalid request (e.g., missing user context)

## Monitoring and Logging

Consider implementing:

1. **Authorization Logging**: Log all authorization decisions
2. **Access Monitoring**: Track who accesses what resources
3. **Permission Auditing**: Regular reviews of user permissions
4. **Security Alerts**: Notify on suspicious access patterns

## Future Enhancements

Potential improvements to the authorization system:

1. **Dynamic Permissions**: Database-driven permissions
2. **Resource-Level Permissions**: Fine-grained resource access control
3. **Time-Based Permissions**: Temporary access grants
4. **Multi-Factor Authorization**: Additional security layers
5. **Permission Inheritance**: Hierarchical permission structures
