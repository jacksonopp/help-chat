package handlers

import (
	"net/http"

	authMiddleware "dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/middleware"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MultiMiddlewareExample demonstrates multiple middlewares on routes
type MultiMiddlewareExample struct {
	ticketService *services.TicketService
}

// NewMultiMiddlewareExample creates a new multi-middleware example handler
func NewMultiMiddlewareExample(ticketService *services.TicketService) *MultiMiddlewareExample {
	return &MultiMiddlewareExample{
		ticketService: ticketService,
	}
}

// RegisterMultiMiddlewareRoutes demonstrates multiple middlewares on routes
func (h *MultiMiddlewareExample) RegisterMultiMiddlewareRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware) {
	// Create route groups
	decoratedGroup := authMiddleware.NewDecoratedRouteGroup(
		e.Group("/api/v1/multi-decorated"),
		authMiddlewareInstance,
	)

	taggedGroup := authMiddleware.NewTaggedRouteGroup(
		e.Group("/api/v1/multi-tagged"),
		authMiddlewareInstance,
	)

	// Example 1: Multiple decorators using NewMultiHandlerDecorator
	decoratedGroup.DELETE("/:id", authMiddleware.NewMultiHandlerDecorator(
		h.DeleteTicketWithMultipleDecorators,
		authMiddleware.AuthorizeAdmin(),                     // Must be admin
		authMiddleware.AuthorizePermission("ticket:delete"), // AND have delete permission
	))

	// Example 2: Multiple decorators using AddDecorator method
	decoratedGroup.PUT("/:id", authMiddleware.NewHandlerDecorator(
		h.UpdateTicketWithMultipleDecorators,
		authMiddleware.Authorize(models.RoleManager, models.RoleAdministrator),
	).AddDecorator(
		authMiddleware.AuthorizePermission("ticket:update"),
	))

	// Example 3: Complex authorization with multiple requirements
	decoratedGroup.POST("/:id/advanced", authMiddleware.NewMultiHandlerDecorator(
		h.AdvancedTicketOperation,
		authMiddleware.AuthorizeAdmin(),                                 // Must be admin
		authMiddleware.AuthorizePermission("ticket:admin"),              // AND have admin permission
		authMiddleware.AuthorizePermission("ticket:advanced:operation"), // AND have advanced operation permission
	))

	// Example 4: Tag-based multiple middlewares using NewMultiTaggedHandler
	taggedGroup.DELETE("/:id", authMiddleware.NewMultiTaggedHandler(
		h.DeleteTicketWithMultipleTags,
		authMiddleware.ParseAuthorizeTag("ADMIN"),         // Must be admin
		authMiddleware.ParseAuthorizeTag("ticket:delete"), // AND have delete permission
	))

	// Example 5: Tag-based using AddTag method
	taggedGroup.PUT("/:id", authMiddleware.NewTaggedHandler(
		h.UpdateTicketWithMultipleTags,
		authMiddleware.ParseAuthorizeTag("MANAGER,ADMIN"),
	).AddTag(
		authMiddleware.ParseAuthorizeTag("ticket:update"),
	))

	// Example 6: Complex tag-based authorization
	taggedGroup.POST("/:id/advanced", authMiddleware.NewMultiTaggedHandler(
		h.AdvancedTicketOperationTagged,
		authMiddleware.ParseAuthorizeTag("ADMIN"),
		authMiddleware.ParseAuthorizeTag("ticket:admin"),
		authMiddleware.ParseAuthorizeTag("Policy=ticket:advanced:operation"),
	))

	// Example 7: Mixed decorators and tags
	decoratedGroup.GET("/:id/sensitive", authMiddleware.NewMultiHandlerDecorator(
		h.GetSensitiveTicketData,
		authMiddleware.AuthorizeAgent(),                             // Must be agent or higher
		authMiddleware.AuthorizePermission("ticket:read:sensitive"), // AND have sensitive read permission
		authMiddleware.AuthorizePermission("data:privacy"),          // AND have data privacy permission
	))

	// Example 8: Conditional authorization based on multiple factors
	taggedGroup.POST("/:id/conditional", authMiddleware.NewMultiTaggedHandler(
		h.ConditionalTicketOperation,
		authMiddleware.ParseAuthorizeTag("AGENT,ADMIN,MANAGER"),          // Must be agent or higher
		authMiddleware.ParseAuthorizeTag("ticket:conditional:operation"), // AND have conditional operation permission
		authMiddleware.ParseAuthorizeTag("Policy=time:business:hours"),   // AND be within business hours
	))
}

// DeleteTicketWithMultipleDecorators - Requires admin AND delete permission
func (h *MultiMiddlewareExample) DeleteTicketWithMultipleDecorators(c echo.Context) error {
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

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket deleted successfully (admin + delete permission required)",
	})
}

// UpdateTicketWithMultipleDecorators - Requires manager/admin AND update permission
func (h *MultiMiddlewareExample) UpdateTicketWithMultipleDecorators(c echo.Context) error {
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

	_, err = h.ticketService.UpdateTicket(c.Request().Context(), ticketID, &req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket updated successfully (manager/admin + update permission required)",
	})
}

// AdvancedTicketOperation - Requires admin AND admin permission AND advanced operation permission
func (h *MultiMiddlewareExample) AdvancedTicketOperation(c echo.Context) error {
	_, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Advanced operation completed (admin + admin permission + advanced operation permission required)",
	})
}

// DeleteTicketWithMultipleTags - Tag-based version
func (h *MultiMiddlewareExample) DeleteTicketWithMultipleTags(c echo.Context) error {
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

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket deleted successfully (ADMIN + ticket:delete tags required)",
	})
}

// UpdateTicketWithMultipleTags - Tag-based version
func (h *MultiMiddlewareExample) UpdateTicketWithMultipleTags(c echo.Context) error {
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

	_, err = h.ticketService.UpdateTicket(c.Request().Context(), ticketID, &req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Ticket updated successfully (MANAGER/ADMIN + ticket:update tags required)",
	})
}

// AdvancedTicketOperationTagged - Tag-based version
func (h *MultiMiddlewareExample) AdvancedTicketOperationTagged(c echo.Context) error {
	_, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Advanced operation completed (ADMIN + ticket:admin + Policy=ticket:advanced:operation tags required)",
	})
}

// GetSensitiveTicketData - Requires agent+ AND sensitive read AND data privacy permissions
func (h *MultiMiddlewareExample) GetSensitiveTicketData(c echo.Context) error {
	_, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Sensitive data retrieved (agent+ + ticket:read:sensitive + data:privacy permissions required)",
	})
}

// ConditionalTicketOperation - Requires agent+ AND conditional operation AND business hours policy
func (h *MultiMiddlewareExample) ConditionalTicketOperation(c echo.Context) error {
	_, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid ticket ID",
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Status:  "success",
		Message: "Conditional operation completed (agent+ + ticket:conditional:operation + Policy=time:business:hours tags required)",
	})
}
