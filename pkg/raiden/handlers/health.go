package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterHealthHandler(e *echo.Echo) {
	e.GET("/health", HealthHandler())
}

func HealthHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := map[string]any{
			"message": "server up",
		}
		return c.JSON(http.StatusOK, data)
	}
}
