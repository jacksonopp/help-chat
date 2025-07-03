package handlers

import (
	authMiddleware "dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/middleware"
	"github.com/labstack/echo/v4"
)

// RouteRegistrar defines the interface for handlers that can register their own routes
type RouteRegistrar interface {
	RegisterRoutes(e *echo.Echo, authMiddlewareInstance *authMiddleware.AuthMiddleware)
}

// SimpleRouteRegistrar defines the interface for handlers that only need the Echo instance
type SimpleRouteRegistrar interface {
	RegisterRoutes(e *echo.Echo)
}
