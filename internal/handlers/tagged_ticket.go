package handlers

import (
	"net/http"

	authMiddleware "dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/middleware"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TaggedTicketHandler demonstrates tag-based authorization (similar to .NET [Authorize])
type TaggedTicketHandler struct {
	ticketService *services.TicketService
}

// NewTaggedTicketHandler creates a new tagged ticket handler
func NewTaggedTicketHandler(ticketService *services.TicketService) *TaggedTicketHandler {
	return &TaggedTicketHandler{
		ticketService: ticketService,
	}
}

// RegisterTaggedRoutes demonstrates tag-based authorization
func (h *TaggedTicketHandler) RegisterTaggedRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
	// Create a tagged route group
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

	// Example 7: No authorization required (public within authenticated group)
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

// DeleteTicketTagged - Admin only
func (h *TaggedTicketHandler) DeleteTicketTagged(c echo.Context) error {
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

// AssignTicketTagged - Agent only
func (h *TaggedTicketHandler) AssignTicketTagged(c echo.Context) error {
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

// GetTicketStatsTagged - Permission-based
func (h *TaggedTicketHandler) GetTicketStatsTagged(c echo.Context) error {
	stats, err := h.ticketService.GetTicketStats(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// EscalateTicketTagged - Multiple permissions
func (h *TaggedTicketHandler) EscalateTicketTagged(c echo.Context) error {
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

// UpdateTicketTagged - Mixed roles and permissions
func (h *TaggedTicketHandler) UpdateTicketTagged(c echo.Context) error {
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

// UpdateTicketStatusTagged - Policy-based
func (h *TaggedTicketHandler) UpdateTicketStatusTagged(c echo.Context) error {
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

// GetTicketTagged - No additional authorization required
func (h *TaggedTicketHandler) GetTicketTagged(c echo.Context) error {
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

// GetMyTicketsTagged - Own tickets permission
func (h *TaggedTicketHandler) GetMyTicketsTagged(c echo.Context) error {
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

// AddCommentTagged - Complex authorization
func (h *TaggedTicketHandler) AddCommentTagged(c echo.Context) error {
	// This is a placeholder for adding comments
	// In a real implementation, you would have a comment service
	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Comment added successfully",
	})
}
