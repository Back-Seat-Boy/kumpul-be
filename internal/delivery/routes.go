package delivery

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *APIHandler) Routes(e *echo.Echo) {
	e.GET("/ping/", h.PingHandler)

	e.GET("/auth/google/login/", h.GoogleLogin)
	e.GET("/auth/google/callback/", h.GoogleCallback)

	api := e.Group("/api")
	api.Use(h.AuthMiddleware())
	{
		api.POST("/auth/logout/", h.Logout)
		api.GET("/me/", h.GetMe)

		api.GET("/users/", h.ListUsers)
		api.GET("/users/:id/", h.GetUserByID)
		api.PUT("/users/:id/", h.UpdateUser)
		api.DELETE("/users/:id/", h.DeleteUser)
	}

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, successResponse("Welcome to kumpul-be API", map[string]string{
			"message": "kumpul-be API",
		}))
	})
}
