package handlers

import (
	"net/http"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/pkg/database"
	"github.com/labstack/echo/v4"
)

// PingHandler handles ping-related requests
type PingHandler struct {
	db *database.Database
}

// NewPingHandler creates a new ping handler
func NewPingHandler(db *database.Database) *PingHandler {
	return &PingHandler{db: db}
}

// RegisterRoutes registers all ping-related routes
func (h *PingHandler) RegisterRoutes(e *echo.Echo) {
	// Ping routes
	e.GET("/ping", h.Ping)
	e.GET("/ping-through", h.PingThrough)
}

// Ping handles the /ping endpoint
// @Summary Health check endpoint
// @Description Simple health check to verify the API is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.PingResponse
// @Router /ping [get]
func (h *PingHandler) Ping(c echo.Context) error {
	response := models.PingResponse{
		Status:  "ok",
		Message: "pong",
	}
	return c.JSON(http.StatusOK, response)
}

// PingThrough handles the /ping-through endpoint
// @Summary Database health check endpoint
// @Description Health check that verifies both the API and database are running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.PingResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /ping-through [get]
func (h *PingHandler) PingThrough(c echo.Context) error {
	if err := h.db.Ping(); err != nil {
		response := models.NewErrorResponse("database unreachable")
		return c.JSON(http.StatusInternalServerError, response)
	}

	response := models.PingResponse{
		Status:  "ok",
		Message: "database pong",
	}
	return c.JSON(http.StatusOK, response)
}
