# Multiple Middleware Authorization

This document describes how to use multiple authorization middlewares on a single route, similar to chaining multiple `[Authorize]` attributes in .NET.

## Overview

The multiple middleware system allows you to apply **multiple authorization requirements** to a single route, where **ALL requirements must be satisfied** (AND logic) for the request to proceed. This is similar to .NET's multiple `[Authorize]` attributes.

## Key Features

- **AND Logic**: All middlewares must pass for access to be granted
- **Flexible Combinations**: Mix roles, permissions, and policies
- **Chaining Support**: Add middlewares incrementally
- **Both Approaches**: Function decorators and tag-based
- **Clear Error Messages**: Specific feedback on which requirement failed

## Function Decorators Approach

### Basic Multiple Decorators

```go
// Example 1: Multiple decorators using NewMultiHandlerDecorator
decoratedGroup.DELETE("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.DeleteTicketWithMultipleDecorators,
    authMiddleware.AuthorizeAdmin(),                    // Must be admin
    authMiddleware.AuthorizePermission("ticket:delete"), // AND have delete permission
))
```

### Chaining Decorators

```go
// Example 2: Multiple decorators using AddDecorator method
decoratedGroup.PUT("/:id", authMiddleware.NewHandlerDecorator(
    h.UpdateTicketWithMultipleDecorators,
    authMiddleware.Authorize(models.RoleManager, models.RoleAdministrator),
).AddDecorator(
    authMiddleware.AuthorizePermission("ticket:update"),
))
```

### Complex Authorization

```go
// Example 3: Complex authorization with multiple requirements
decoratedGroup.POST("/:id/advanced", authMiddleware.NewMultiHandlerDecorator(
    h.AdvancedTicketOperation,
    authMiddleware.AuthorizeAdmin(),                    // Must be admin
    authMiddleware.AuthorizePermission("ticket:admin"), // AND have admin permission
    authMiddleware.AuthorizePermission("ticket:advanced:operation"), // AND have advanced operation permission
))
```

## Tag-Based Approach

### Basic Multiple Tags

```go
// Example 4: Tag-based multiple middlewares using NewMultiTaggedHandler
taggedGroup.DELETE("/:id", authMiddleware.NewMultiTaggedHandler(
    h.DeleteTicketWithMultipleTags,
    authMiddleware.ParseAuthorizeTag("ADMIN"),           // Must be admin
    authMiddleware.ParseAuthorizeTag("ticket:delete"),   // AND have delete permission
))
```

### Chaining Tags

```go
// Example 5: Tag-based using AddTag method
taggedGroup.PUT("/:id", authMiddleware.NewTaggedHandler(
    h.UpdateTicketWithMultipleTags,
    authMiddleware.ParseAuthorizeTag("MANAGER,ADMIN"),
).AddTag(
    authMiddleware.ParseAuthorizeTag("ticket:update"),
))
```

### Complex Tag-Based Authorization

```go
// Example 6: Complex tag-based authorization
taggedGroup.POST("/:id/advanced", authMiddleware.NewMultiTaggedHandler(
    h.AdvancedTicketOperationTagged,
    authMiddleware.ParseAuthorizeTag("ADMIN"),
    authMiddleware.ParseAuthorizeTag("ticket:admin"),
    authMiddleware.ParseAuthorizeTag("Policy=ticket:advanced:operation"),
))
```

## Advanced Examples

### Sensitive Data Access

```go
// Example 7: Multiple permissions for sensitive data
decoratedGroup.GET("/:id/sensitive", authMiddleware.NewMultiHandlerDecorator(
    h.GetSensitiveTicketData,
    authMiddleware.AuthorizeAgent(),                     // Must be agent or higher
    authMiddleware.AuthorizePermission("ticket:read:sensitive"), // AND have sensitive read permission
    authMiddleware.AuthorizePermission("data:privacy"),  // AND have data privacy permission
))
```

### Conditional Operations

```go
// Example 8: Conditional authorization based on multiple factors
taggedGroup.POST("/:id/conditional", authMiddleware.NewMultiTaggedHandler(
    h.ConditionalTicketOperation,
    authMiddleware.ParseAuthorizeTag("AGENT,ADMIN,MANAGER"), // Must be agent or higher
    authMiddleware.ParseAuthorizeTag("ticket:conditional:operation"), // AND have conditional operation permission
    authMiddleware.ParseAuthorizeTag("Policy=time:business:hours"), // AND be within business hours
))
```

## Comparison with .NET

### .NET ASP.NET Core

```csharp
[Authorize(Roles = "Admin")]
[Authorize(Policy = "TicketDelete")]
[HttpDelete("{id}")]
public IActionResult DeleteTicket(int id) { ... }

[Authorize(Roles = "Admin,Manager")]
[Authorize(Policy = "TicketUpdate")]
[Authorize(Policy = "BusinessHours")]
[HttpPut("{id}")]
public IActionResult UpdateTicket(int id) { ... }
```

### Go Function Decorators

```go
// Multiple decorators (AND logic)
decoratedGroup.DELETE("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.DeleteTicketWithMultipleDecorators,
    authMiddleware.AuthorizeAdmin(),
    authMiddleware.AuthorizePermission("ticket:delete"),
))

decoratedGroup.PUT("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.UpdateTicketWithMultipleDecorators,
    authMiddleware.Authorize(models.RoleAdministrator, models.RoleManager),
    authMiddleware.AuthorizePermission("ticket:update"),
    authMiddleware.AuthorizePermission("business:hours"),
))
```

### Go Tag-Based

```go
// Multiple tags (AND logic)
taggedGroup.DELETE("/:id", authMiddleware.NewMultiTaggedHandler(
    h.DeleteTicketWithMultipleTags,
    authMiddleware.ParseAuthorizeTag("ADMIN"),
    authMiddleware.ParseAuthorizeTag("Policy=ticket:delete"),
))

taggedGroup.PUT("/:id", authMiddleware.NewMultiTaggedHandler(
    h.UpdateTicketWithMultipleTags,
    authMiddleware.ParseAuthorizeTag("ADMIN,MANAGER"),
    authMiddleware.ParseAuthorizeTag("Policy=ticket:update"),
    authMiddleware.ParseAuthorizeTag("Policy=business:hours"),
))
```

## Available Methods

### Function Decorators

#### Constructor Methods

- `NewMultiHandlerDecorator(handler, decorators...)` - Create with multiple decorators
- `NewHandlerDecorator(handler, decorator)` - Create with single decorator

#### Chaining Methods

- `AddDecorator(decorator)` - Add additional decorator to existing handler

### Tag-Based

#### Constructor Methods

- `NewMultiTaggedHandler(handler, tags...)` - Create with multiple tags
- `NewTaggedHandler(handler, tag)` - Create with single tag

#### Chaining Methods

- `AddTag(tag)` - Add additional tag to existing handler

#### Parsing Methods

- `ParseMultipleAuthorizeTags(tags...)` - Parse multiple authorization tags

## Authorization Logic

### AND Logic (All Must Pass)

All middlewares must pass for access to be granted:

```go
// User must satisfy ALL requirements:
// 1. Be an ADMIN OR MANAGER
// 2. AND have ticket:update permission
// 3. AND have business:hours permission
decoratedGroup.PUT("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.UpdateTicket,
    authMiddleware.Authorize(models.RoleAdministrator, models.RoleManager),
    authMiddleware.AuthorizePermission("ticket:update"),
    authMiddleware.AuthorizePermission("business:hours"),
))
```

### Evaluation Order

1. **Roles are checked first** - If user has any of the required roles, that decorator passes
2. **Permissions are checked second** - If user has any of the required permissions, that decorator passes
3. **Policies are checked last** - If user has the required policy, that decorator passes
4. **All decorators must pass** - If any decorator fails, access is denied

### Error Handling

- **401 Unauthorized** - User not found in context
- **403 Forbidden** - User doesn't meet authorization requirements
- **Clear feedback** - Specific error messages for debugging

## Best Practices

### 1. **Logical Grouping**

```go
// Good: Group related requirements
decoratedGroup.POST("/:id/advanced", authMiddleware.NewMultiHandlerDecorator(
    h.AdvancedOperation,
    authMiddleware.AuthorizeAdmin(),                    // Role requirement
    authMiddleware.AuthorizePermission("ticket:admin"), // Permission requirement
    authMiddleware.AuthorizePermission("ticket:advanced:operation"), // Specific operation permission
))
```

### 2. **Meaningful Permissions**

```go
// Good: Specific, descriptive permissions
authMiddleware.AuthorizePermission("ticket:delete:sensitive")
authMiddleware.AuthorizePermission("data:privacy:access")

// Avoid: Generic permissions
authMiddleware.AuthorizePermission("delete")
authMiddleware.AuthorizePermission("access")
```

### 3. **Documentation**

```go
// Add comments to explain complex authorization
// Requires admin role AND delete permission AND business hours policy
decoratedGroup.DELETE("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.DeleteTicket,
    authMiddleware.AuthorizeAdmin(),
    authMiddleware.AuthorizePermission("ticket:delete"),
    authMiddleware.AuthorizePermission("business:hours"),
))
```

### 4. **Performance Considerations**

```go
// Good: Check roles first (faster)
decoratedGroup.GET("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.GetTicket,
    authMiddleware.AuthorizeAgent(), // Fast role check
    authMiddleware.AuthorizePermission("ticket:read:sensitive"), // Slower permission check
))

// Avoid: Too many permission checks
decoratedGroup.GET("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.GetTicket,
    authMiddleware.AuthorizePermission("permission1"),
    authMiddleware.AuthorizePermission("permission2"),
    authMiddleware.AuthorizePermission("permission3"),
    authMiddleware.AuthorizePermission("permission4"),
    authMiddleware.AuthorizePermission("permission5"),
))
```

## Migration Guide

### From Single Middleware

**Before:**

```go
// Single middleware
decoratedGroup.DELETE("/:id", authMiddleware.NewHandlerDecorator(
    h.DeleteTicket,
    authMiddleware.AuthorizeAdmin(),
))
```

**After:**

```go
// Multiple middlewares
decoratedGroup.DELETE("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.DeleteTicket,
    authMiddleware.AuthorizeAdmin(),
    authMiddleware.AuthorizePermission("ticket:delete"),
))
```

### From Traditional Middleware

**Before:**

```go
// Traditional middleware approach
tickets.DELETE("/:id", h.DeleteTicket,
    authMiddlewareInstance.RequireAdmin(),
    authMiddlewareInstance.RequirePermission("ticket:delete"),
)
```

**After:**

```go
// Decorator approach
decoratedGroup.DELETE("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.DeleteTicket,
    authMiddleware.AuthorizeAdmin(),
    authMiddleware.AuthorizePermission("ticket:delete"),
))
```

## Testing Multiple Middlewares

### Test Cases

```go
func TestMultipleMiddlewareAuthorization(t *testing.T) {
    testCases := []struct {
        name           string
        user           *models.User
        expectedStatus int
        description    string
    }{
        {
            name:           "Admin with all permissions can access",
            user:           createTestUser(models.RoleAdministrator),
            expectedStatus: 200,
            description:    "Administrator with all required permissions should access",
        },
        {
            name:           "Manager without delete permission cannot access",
            user:           createTestUser(models.RoleManager),
            expectedStatus: 403,
            description:    "Manager without delete permission should be denied",
        },
        {
            name:           "Agent cannot access admin-only route",
            user:           createTestUser(models.RoleSupportAgent),
            expectedStatus: 403,
            description:    "Agent should be denied access to admin-only route",
        },
    }

    // Test implementation
}
```

## Common Patterns

### 1. **Role + Permission Pattern**

```go
// Common pattern: Require specific role AND specific permission
decoratedGroup.DELETE("/:id", authMiddleware.NewMultiHandlerDecorator(
    h.DeleteTicket,
    authMiddleware.AuthorizeAdmin(),                    // Role requirement
    authMiddleware.AuthorizePermission("ticket:delete"), // Permission requirement
))
```

### 2. **Multiple Permission Pattern**

```go
// Require multiple permissions for sensitive operations
decoratedGroup.GET("/:id/sensitive", authMiddleware.NewMultiHandlerDecorator(
    h.GetSensitiveData,
    authMiddleware.AuthorizePermission("ticket:read:sensitive"),
    authMiddleware.AuthorizePermission("data:privacy"),
    authMiddleware.AuthorizePermission("audit:access"),
))
```

### 3. **Conditional Pattern**

```go
// Require role + permission + policy (time-based, location-based, etc.)
taggedGroup.POST("/:id/conditional", authMiddleware.NewMultiTaggedHandler(
    h.ConditionalOperation,
    authMiddleware.ParseAuthorizeTag("AGENT,ADMIN,MANAGER"),
    authMiddleware.ParseAuthorizeTag("ticket:conditional:operation"),
    authMiddleware.ParseAuthorizeTag("Policy=time:business:hours"),
))
```

## Conclusion

The multiple middleware system provides a powerful and flexible way to implement complex authorization requirements. It offers:

- **AND Logic**: All requirements must be satisfied
- **Flexible Combinations**: Mix roles, permissions, and policies
- **Chaining Support**: Add middlewares incrementally
- **Clear Error Handling**: Specific feedback on failures
- **Performance Optimized**: Efficient evaluation order
- **Familiar Syntax**: Similar to .NET's multiple `[Authorize]` attributes

This system allows you to implement sophisticated authorization patterns while maintaining code clarity and performance.
