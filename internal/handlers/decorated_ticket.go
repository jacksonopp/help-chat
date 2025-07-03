package handlers

import (
	"net/http"

	authMiddleware "dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/middleware"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DecoratedTicketHandler demonstrates decorator-style authorization
type DecoratedTicketHandler struct {
	ticketService *services.TicketService
}

// NewDecoratedTicketHandler creates a new decorated ticket handler
func NewDecoratedTicketHandler(ticketService *services.TicketService) *DecoratedTicketHandler {
	return &DecoratedTicketHandler{
		ticketService: ticketService,
	}
}

// RegisterDecoratedRoutes demonstrates decorator-style authorization
func (h *DecoratedTicketHandler) RegisterDecoratedRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
	// Create a decorated route group
	tickets := e.Group("/api/v1/decorated-tickets")
	tickets.Use(authMiddlewareInstance.Authenticate)

	decoratedGroup := authMiddleware.NewDecoratedRouteGroup(tickets, authMiddlewareInstance)

	// Example 1: Admin-only route using AuthorizeAdmin decorator
	decoratedGroup.DELETE("/:id", authMiddleware.NewHandlerDecorator(
		h.DeleteTicketDecorated,
		authMiddleware.AuthorizeAdmin(),
	))

	// Example 2: Agent-only route using AuthorizeAgent decorator
	decoratedGroup.POST("/:id/assign", authMiddleware.NewHandlerDecorator(
		h.AssignTicketDecorated,
		authMiddleware.AuthorizeAgent(),
	))

	// Example 3: Specific roles using Authorize decorator
	decoratedGroup.POST("/:id/status", authMiddleware.NewHandlerDecorator(
		h.UpdateTicketStatusDecorated,
		authMiddleware.Authorize(models.RoleSupportAgent, models.RoleManager, models.RoleAdministrator),
	))

	// Example 4: Permission-based authorization
	decoratedGroup.GET("/stats", authMiddleware.NewHandlerDecorator(
		h.GetTicketStatsDecorated,
		authMiddleware.AuthorizePermission("ticket:stats:read"),
	))

	// Example 5: Multiple permissions (any of them)
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

	// Example 7: No authorization required (public within authenticated group)
	decoratedGroup.GET("/:id", authMiddleware.NewHandlerDecorator(
		h.GetTicketDecorated,
		nil, // No decorator means no additional authorization required
	))

	// Example 8: Custom permission check
	decoratedGroup.GET("/my", authMiddleware.NewHandlerDecorator(
		h.GetMyTicketsDecorated,
		authMiddleware.AuthorizePermission("ticket:read:own"),
	))
}

// DeleteTicketDecorated - Admin only
func (h *DecoratedTicketHandler) DeleteTicketDecorated(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}

	err = h.ticketService.DeleteTicket(c.Request().Context(), ticketID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// AssignTicketDecorated - Agent only
func (h *DecoratedTicketHandler) AssignTicketDecorated(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	var req models.AssignTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}

	err = h.ticketService.AssignTicket(c.Request().Context(), ticketID, req.AgentID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket assigned successfully",
	})
}

// UpdateTicketStatusDecorated - Specific roles
func (h *DecoratedTicketHandler) UpdateTicketStatusDecorated(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	var req models.UpdateTicketStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}

	err = h.ticketService.UpdateTicketStatus(c.Request().Context(), ticketID, &req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket status updated successfully",
	})
}

// GetTicketStatsDecorated - Permission-based
func (h *DecoratedTicketHandler) GetTicketStatsDecorated(c echo.Context) error {
	stats, err := h.ticketService.GetTicketStats(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// EscalateTicketDecorated - Multiple permissions
func (h *DecoratedTicketHandler) EscalateTicketDecorated(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	var req models.EscalateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}

	err = h.ticketService.EscalateTicket(c.Request().Context(), ticketID, &req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket escalated successfully",
	})
}

// UpdateTicketDecorated - Mixed roles and permissions
func (h *DecoratedTicketHandler) UpdateTicketDecorated(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	var req models.UpdateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}

	ticket, err := h.ticketService.UpdateTicket(c.Request().Context(), ticketID, &req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, ticket)
}

// GetTicketDecorated - No additional authorization required
func (h *DecoratedTicketHandler) GetTicketDecorated(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	ticket, err := h.ticketService.GetTicket(c.Request().Context(), ticketID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if ticket == nil {
		return c.JSON(http.StatusNotFound, models.ErrorResponse{
			Status:  "error",
			Message: "Ticket not found",
		})
	}

	return c.JSON(http.StatusOK, ticket)
}

// GetMyTicketsDecorated - Own tickets permission
func (h *DecoratedTicketHandler) GetMyTicketsDecorated(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}

	query := buildTicketQueryFromRequest(c)
	tickets, err := h.ticketService.GetTicketsByUser(c.Request().Context(), userID, query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, tickets)
}
