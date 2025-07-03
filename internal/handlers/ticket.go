package handlers

import (
	"net/http"
	"strconv"

	authMiddleware "dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/middleware"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TicketHandler handles ticket-related HTTP requests
type TicketHandler struct {
	ticketService *services.TicketService
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(ticketService *services.TicketService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
	}
}

// RegisterRoutes registers the ticket routes
func (h *TicketHandler) RegisterRoutes(e *echo.Echo, ami *authMiddleware.AuthMiddleware) {
	// Public routes (require authentication)
	tickets := e.Group("/api/v1/tickets")
	tickets.Use(ami.Authenticate)

	// // Ticket CRUD operations
	tickets.GET("", h.ListTickets, ami.RequireAgent(), ami.RequireManager(), ami.RequireAdmin())
	tickets.POST("", h.CreateTicket)
	tickets.GET("/:id", h.GetTicket, ami.RequireAnyRole(models.RoleSupportAgent, models.RoleManager, models.RoleAdministrator), ami.RequireOwnerOrAdmin(func(c echo.Context) (string, error) {
		return h.getUserId(c)
	}))
	tickets.PUT("/:id", h.UpdateTicket)
	tickets.DELETE("/:id", h.DeleteTicket, ami.RequireAdmin()) // Admin only

	// Ticket actions - require agent or admin privileges
	tickets.POST("/:id/assign", h.AssignTicket, ami.RequireAgent())
	tickets.POST("/:id/status", h.UpdateTicketStatus, ami.RequireAgent())
	tickets.POST("/:id/escalate", h.EscalateTicket, ami.RequireAgent())

	// User-specific routes
	tickets.GET("/my", h.GetMyTickets)
	tickets.GET("/assigned", h.GetAssignedTickets)

	// Statistics - require agent or admin privileges
	tickets.GET("/stats", h.GetTicketStats, ami.RequireAgent())
}

// CreateTicket handles ticket creation
// @Summary Create a new ticket
// @Description Create a new support ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Param ticket body models.CreateTicketRequest true "Ticket data"
// @Success 201 {object} models.Ticket
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets [post]
// @Security ApiKeyAuth
func (h *TicketHandler) CreateTicket(c echo.Context) error {
	var req models.CreateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponseFromError(err))
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
	}

	ticket, err := h.ticketService.CreateTicket(c.Request().Context(), &req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.JSON(http.StatusCreated, ticket)
}

// GetTicket handles retrieving a single ticket
// @Summary Get a ticket by ID
// @Description Retrieve a ticket by its ID
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Success 200 {object} models.Ticket
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets/{id} [get]
// @Security ApiKeyAuth
func (h *TicketHandler) GetTicket(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid ticket ID"))
	}

	ticket, err := h.ticketService.GetTicket(c.Request().Context(), ticketID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	if ticket == nil {
		return c.JSON(http.StatusNotFound, models.NewErrorResponse("Ticket not found"))
	}

	return c.JSON(http.StatusOK, ticket)
}

// UpdateTicket handles ticket updates
// @Summary Update a ticket
// @Description Update an existing ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param ticket body models.UpdateTicketRequest true "Updated ticket data"
// @Success 200 {object} models.Ticket
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets/{id} [put]
// @Security ApiKeyAuth
func (h *TicketHandler) UpdateTicket(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid ticket ID"))
	}

	var req models.UpdateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponseFromError(err))
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
	}

	ticket, err := h.ticketService.UpdateTicket(c.Request().Context(), ticketID, &req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.JSON(http.StatusOK, ticket)
}

// DeleteTicket handles ticket deletion
// @Summary Delete a ticket
// @Description Delete a ticket (admin only)
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets/{id} [delete]
// @Security ApiKeyAuth
func (h *TicketHandler) DeleteTicket(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid ticket ID"))
	}

	// Get user ID from context for authorization
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
	}

	err = h.ticketService.DeleteTicket(c.Request().Context(), ticketID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.NoContent(http.StatusNoContent)
}

// ListTickets handles listing tickets with filtering and pagination
// @Summary List tickets
// @Description Retrieve a list of tickets with filtering and pagination
// @Tags tickets
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param category_id query string false "Filter by category ID"
// @Param assigned_to query string false "Filter by assigned agent ID"
// @Param created_by query string false "Filter by creator ID"
// @Param search query string false "Search in title and description"
// @Success 200 {object} models.TicketListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets [get]
// @Security ApiKeyAuth
func (h *TicketHandler) ListTickets(c echo.Context) error {
	query := &models.TicketQuery{
		Page:     1,
		PageSize: 20,
	}

	// Parse pagination parameters
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}

	if pageSizeStr := c.QueryParam("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			query.PageSize = pageSize
		}
	}

	// Parse filter parameters
	filter := &models.TicketFilter{}

	if status := c.QueryParam("status"); status != "" {
		ticketStatus := models.TicketStatus(status)
		filter.Status = &ticketStatus
	}

	if priority := c.QueryParam("priority"); priority != "" {
		ticketPriority := models.TicketPriority(priority)
		filter.Priority = &ticketPriority
	}

	if categoryIDStr := c.QueryParam("category_id"); categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			filter.CategoryID = &categoryID
		}
	}

	if assignedToStr := c.QueryParam("assigned_to"); assignedToStr != "" {
		if assignedTo, err := uuid.Parse(assignedToStr); err == nil {
			filter.AssignedTo = &assignedTo
		}
	}

	if createdByStr := c.QueryParam("created_by"); createdByStr != "" {
		if createdBy, err := uuid.Parse(createdByStr); err == nil {
			filter.CreatedBy = &createdBy
		}
	}

	if search := c.QueryParam("search"); search != "" {
		filter.Search = search
	}

	query.Filter = filter

	// Parse sorting parameters
	if sortField := c.QueryParam("sort_field"); sortField != "" {
		if sortDirection := c.QueryParam("sort_direction"); sortDirection != "" {
			query.Sort = &models.TicketSort{
				Field:     sortField,
				Direction: sortDirection,
			}
		}
	}

	tickets, err := h.ticketService.ListTickets(c.Request().Context(), query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.JSON(http.StatusOK, tickets)
}

// AssignTicket handles ticket assignment
// @Summary Assign a ticket to an agent
// @Description Assign a ticket to a support agent
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param assignment body models.AssignTicketRequest true "Assignment data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets/{id}/assign [post]
// @Security ApiKeyAuth
func (h *TicketHandler) AssignTicket(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid ticket ID"))
	}

	var req models.AssignTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponseFromError(err))
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
	}

	err = h.ticketService.AssignTicket(c.Request().Context(), ticketID, req.AgentID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket assigned successfully",
	})
}

// UpdateTicketStatus handles ticket status updates
// @Summary Update ticket status
// @Description Update the status of a ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param status body models.UpdateTicketStatusRequest true "Status update data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets/{id}/status [post]
// @Security ApiKeyAuth
func (h *TicketHandler) UpdateTicketStatus(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid ticket ID"))
	}

	var req models.UpdateTicketStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponseFromError(err))
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
	}

	err = h.ticketService.UpdateTicketStatus(c.Request().Context(), ticketID, &req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket status updated successfully",
	})
}

// EscalateTicket handles ticket escalation
// @Summary Escalate a ticket
// @Description Escalate a ticket to a manager or administrator
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param escalation body models.EscalateTicketRequest true "Escalation data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets/{id}/escalate [post]
// @Security ApiKeyAuth
func (h *TicketHandler) EscalateTicket(c echo.Context) error {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid ticket ID"))
	}

	var req models.EscalateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorResponseFromError(err))
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
	}

	err = h.ticketService.EscalateTicket(c.Request().Context(), ticketID, &req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket escalated successfully",
	})
}

// GetMyTickets handles retrieving tickets created by the current user
// @Summary Get my tickets
// @Description Retrieve tickets created by the current user
// @Tags tickets
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} models.TicketListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets/my [get]
// @Security ApiKeyAuth
func (h *TicketHandler) GetMyTickets(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
	}

	query := buildTicketQueryFromRequest(c)
	tickets, err := h.ticketService.GetTicketsByUser(c.Request().Context(), userID, query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.JSON(http.StatusOK, tickets)
}

// GetAssignedTickets handles retrieving tickets assigned to the current user
// @Summary Get assigned tickets
// @Description Retrieve tickets assigned to the current user
// @Tags tickets
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} models.TicketListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets/assigned [get]
// @Security ApiKeyAuth
func (h *TicketHandler) GetAssignedTickets(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.NewErrorResponse("Unauthorized"))
	}

	query := buildTicketQueryFromRequest(c)
	tickets, err := h.ticketService.GetTicketsByAgent(c.Request().Context(), userID, query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.JSON(http.StatusOK, tickets)
}

// GetTicketStats handles retrieving ticket statistics
// @Summary Get ticket statistics
// @Description Retrieve ticket statistics
// @Tags tickets
// @Accept json
// @Produce json
// @Success 200 {object} models.TicketStats
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/tickets/stats [get]
// @Security ApiKeyAuth
func (h *TicketHandler) GetTicketStats(c echo.Context) error {
	stats, err := h.ticketService.GetTicketStats(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorResponseFromError(err))
	}

	return c.JSON(http.StatusOK, stats)
}

// Helper functions

func getUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userIDStr := c.Get("user_id").(string)
	if userIDStr == "" {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "user ID not found in context")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID in context")
	}

	return userID, nil
}

func getUserRoleFromContext(c echo.Context) (models.UserRole, error) {
	userRoleStr := c.Get("user_role").(string)
	if userRoleStr == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "user role not found in context")
	}

	userRole := models.UserRole(userRoleStr)

	// Validate the role
	validRoles := []models.UserRole{
		models.RoleEndUser,
		models.RoleSupportAgent,
		models.RoleManager,
		models.RoleAdministrator,
	}

	for _, validRole := range validRoles {
		if userRole == validRole {
			return userRole, nil
		}
	}

	return "", echo.NewHTTPError(http.StatusUnauthorized, "invalid user role in context")
}

func getUserFromContext(c echo.Context) (*models.User, error) {
	user := c.Get("user").(*models.User)
	if user == nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
	}
	return user, nil
}

func buildTicketQueryFromRequest(c echo.Context) *models.TicketQuery {
	query := &models.TicketQuery{
		Page:     1,
		PageSize: 20,
	}

	// Parse pagination parameters
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}

	if pageSizeStr := c.QueryParam("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			query.PageSize = pageSize
		}
	}

	return query
}

// func(c echo.Context) (string, error) {
// 	return h.getUserId(c)
// }

func (h *TicketHandler) getUserId(c echo.Context) (string, error) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return "", err
	}
	ticket, err := h.ticketService.GetTicket(c.Request().Context(), ticketID)
	if err != nil {
		return "", err
	}
	return ticket.CreatedByID.String(), nil
}
