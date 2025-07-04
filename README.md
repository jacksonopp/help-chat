# Help Chat Server

A simple Go HTTP server with endpoints to ping the server and database.

## Features

- **Server Ping Endpoint**: `/ping` - Check if the server is running
- **Database Ping Endpoint**: `/ping-db` - Check if the database is connected and responding
- **Health Check Endpoint**: `/health` - Comprehensive health check for all services
- **Graceful Shutdown**: Proper shutdown handling with signal management
- **Request Logging**: Automatic logging of all HTTP requests with timing
- **CORS Support**: Cross-origin resource sharing headers for web applications
- **Connection Pooling**: Optimized database connection management
- **Timeout Handling**: Context-based timeouts for database operations

## Prerequisites

- Go 1.24.4 or later
- SQLite (included with Go)

## Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Configuration

The server uses environment variables for configuration. You can set these in your environment or create a `.env` file.

### Environment Variables

| Variable  | Default       | Description                      |
| --------- | ------------- | -------------------------------- |
| `PORT`    | `8080`        | Port for the server to listen on |
| `HOST`    | `0.0.0.0`     | Host for the server to bind to   |
| `DB_FILE` | `helpchat.db` | SQLite database file path        |

### Example `.env` file

```env
PORT=8080
HOST=0.0.0.0
DB_FILE=helpchat.db
```

## Running the Server

### Development

```bash
go run cmd/server/main.go
```

### Production

```bash
go build -o helpchat-server cmd/server/main.go
./helpchat-server
```

## API Documentation

The API documentation is available via Swagger UI at `/swagger/index.html` when the server is running.

### Regenerating Documentation

To regenerate the Swagger documentation after making changes to the API:

```bash
make swagger
```

Or manually:

```bash
swag init -g cmd/server/main.go -o docs
```

## API Endpoints

### GET /ping

Ping the server to check if it's running.

**Response:**

```json
{
  "status": "success",
  "message": "Server is running",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### GET /ping-db

Ping the database to check if it's connected and responding.

**Success Response:**

```json
{
  "status": "success",
  "message": "Database is connected and responding",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error Response (if database is not connected):**

```json
{
  "status": "error",
  "message": "Database not connected",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error Response (if database ping fails):**

```json
{
  "status": "error",
  "message": "Database ping failed: [error details]",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### GET /health

Comprehensive health check for all services.

**Success Response:**

```json
{
  "status": "success",
  "message": "Health check completed",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "server": "healthy",
    "database": "healthy"
  }
}
```

**Response with Database Issues:**

```json
{
  "status": "success",
  "message": "Health check completed",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "server": "healthy",
    "database": "unhealthy"
  }
}
```

## Testing the Endpoints

You can test the endpoints using curl:

```bash
# Ping the server
curl http://localhost:8080/ping

# Ping the database
curl http://localhost:8080/ping-db

# Health check for all services
curl http://localhost:8080/health
```

## Database Setup

The server uses SQLite, which is a lightweight, file-based database. The database file will be created automatically when the server starts.

### Database File Location

By default, the database file is created in the current directory as `helpchat.db`. You can change this by setting the `DB_FILE` environment variable:

```bash
export DB_FILE=/path/to/your/database.db
```

### Database File Permissions

Make sure the application has read/write permissions to the directory where the database file will be created.

## Logs

The server provides informative logs about:

- Server startup and port
- Available endpoints
- Database connection status
- Any connection errors
- Request/response timing for all endpoints
- Graceful shutdown process

## Error Handling

- The server gracefully handles database connection failures
- All endpoints return proper HTTP status codes
- JSON responses are consistently formatted with timestamps
- Method validation is implemented for all endpoints
- Graceful shutdown with timeout handling
- Context-based timeouts for database operations
- Comprehensive health checking for all services
