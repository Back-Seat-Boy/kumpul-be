package delivery

import (
	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/labstack/echo/v4"
)

func (h *APIHandler) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			sessionID := extractBearerToken(c)

			if sessionID == "" {
				return model.ErrUnauthorized
			}

			session, user, err := h.authUsecase.ValidateSession(ctx, sessionID)
			if err != nil {
				return model.ErrUnauthorized
			}

			c.Set(string(model.ContextKeySession), session)
			c.Set(string(model.ContextKeyUser), toUserInfo(user))

			return next(c)
		}
	}
}
