package middleware

import (
	"net/http"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"github.com/labstack/echo/v4"
)

// ErrorHandlerMiddleware creates middleware that handles HTTP errors and converts them to standardized error responses
func ErrorHandlerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Call the next handler
			err := next(c)
			if err == nil {
				return nil
			}

			// Check if it's an HTTP error
			if httpError, ok := err.(*echo.HTTPError); ok {
				// Convert the error message to a slice
				var messages []string
				if httpError.Message != nil {
					if msg, ok := httpError.Message.(string); ok {
						messages = []string{msg}
					} else {
						messages = []string{"An error occurred"}
					}
				} else {
					messages = []string{"An error occurred"}
				}

				// Create standardized error response
				errorResponse := models.NewErrorResponseWithMessages("Request failed", messages)

				// Return the error response with the appropriate status code
				return c.JSON(httpError.Code, errorResponse)
			}

			// For other types of errors, return a generic error response
			errorResponse := models.NewErrorResponseFromError(err)
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}
	}
}
