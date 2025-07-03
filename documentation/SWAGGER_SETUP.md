# Swagger Documentation Setup

This document explains how Swagger documentation has been implemented in the HelpChat API.

## Overview

Swagger (OpenAPI) documentation has been integrated into the HelpChat API using the `swaggo/swag` library. This provides interactive API documentation that can be accessed via a web browser.

## Features Implemented

### 1. Swagger UI Integration

- **URL**: `/swagger/index.html`
- **Description**: Interactive web interface for exploring the API
- **Features**:
  - Try out endpoints directly from the browser
  - View request/response schemas
  - See example values
  - Authentication support (Bearer token)

### 2. API Documentation

- **Health Check Endpoints**: Documented with proper response models
- **Response Models**: Structured response types with examples
- **Error Handling**: Proper error response documentation
- **Tags**: Organized endpoints by category (health)

### 3. Generated Files

- `docs/docs.go`: Go package with embedded Swagger spec
- `docs/swagger.json`: JSON format of the API specification
- `docs/swagger.yaml`: YAML format of the API specification

## How to Use

### Starting the Server

```bash
go run cmd/server/main.go
```

### Accessing Swagger UI

1. Start the server
2. Open your browser to: `http://localhost:8080/swagger/index.html`
3. Explore the API endpoints interactively

### Testing Endpoints

You can test endpoints directly from the Swagger UI:

1. Click on an endpoint (e.g., `/ping`)
2. Click "Try it out"
3. Click "Execute"
4. View the response

## Development Workflow

### Adding New Endpoints

1. Create your handler function
2. Add Swagger annotations above the function:

```go
// @Summary Brief description
// @Description Detailed description
// @Tags category
// @Accept json
// @Produce json
// @Success 200 {object} models.ResponseType
// @Failure 400 {object} models.ErrorResponse
// @Router /endpoint [method]
func (h *Handler) Endpoint(c echo.Context) error {
    // Implementation
}
```

### Adding Response Models

1. Create structs in `internal/models/`
2. Add `@Description` comments
3. Use `example` tags for sample values:

```go
// @Description Response description
type ResponseType struct {
    Status  string `json:"status" example:"ok"`
    Message string `json:"message" example:"success"`
}
```

### Regenerating Documentation

After making changes to annotations or models:

```bash
make swagger
```

Or manually:

```bash
swag init -g cmd/server/main.go -o docs
```

## Available Annotations

### Basic Annotations

- `@Summary`: Brief endpoint description
- `@Description`: Detailed endpoint description
- `@Tags`: Category for grouping endpoints
- `@Accept`: Content types the endpoint accepts
- `@Produce`: Content types the endpoint produces
- `@Router`: HTTP method and path

### Response Annotations

- `@Success`: Success response (status code, description, schema)
- `@Failure`: Error response (status code, description, schema)
- `@Response`: Generic response annotation

### Security Annotations

- `@Security`: Security requirements for the endpoint
- `@SecurityDefinitions`: Define security schemes

### Parameter Annotations

- `@Param`: Define parameters (path, query, header, body)
- `@Header`: Define response headers

## Example Endpoint Documentation

```go
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User information"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Router /users [post]
func (h *UserHandler) CreateUser(c echo.Context) error {
    // Implementation
}
```

## Best Practices

1. **Always document responses**: Include both success and error responses
2. **Use meaningful examples**: Provide realistic example values
3. **Group related endpoints**: Use consistent tags
4. **Keep descriptions clear**: Write descriptions that help API consumers
5. **Update documentation**: Regenerate docs after making changes
6. **Use structured models**: Create proper response/request models instead of generic maps

## Troubleshooting

### Common Issues

1. **Import errors**: Make sure the docs package is imported in main.go
2. **Missing annotations**: Check that all required annotations are present
3. **Build errors**: Ensure all response models are properly defined
4. **Swagger UI not loading**: Verify the server is running and the route is configured

### Regenerating from Scratch

If you encounter issues, you can regenerate everything:

```bash
rm -rf docs/
make swagger
```

## Additional Resources

- [Swaggo Documentation](https://github.com/swaggo/swag)
- [Echo Swagger](https://github.com/swaggo/echo-swagger)
- [OpenAPI Specification](https://swagger.io/specification/)
