package delivery

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
)

func (h *APIHandler) GoogleLogin(c echo.Context) error {
	ctx := c.Request().Context()
	url := h.authUsecase.GetGoogleLoginURL(ctx)
	return c.JSON(http.StatusOK, successResponse("Google login URL generated", LoginURLResponse{LoginURL: url}))
}

func (h *APIHandler) GoogleCallback(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")

	sessionID, user, err := h.authUsecase.HandleGoogleCallback(ctx, code)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpOutGoingContext(ctx), "code": code}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Login successful", AuthResponse{
		SessionID: sessionID,
		User:      toUserInfo(user),
	}))
}

func (h *APIHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	sessionID := extractBearerToken(c)

	if sessionID != "" {
		err := h.authUsecase.Logout(ctx, sessionID)
		if err != nil {
			log.WithFields(log.Fields{"context": utils.DumpOutGoingContext(ctx), "sessionID": sessionID}).Error()
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, successResponse("Logout successful", nil))
}

func (h *APIHandler) GetMe(c echo.Context) error {
	user := c.Get(string(model.ContextKeyUser))
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, model.ErrUnauthorized.Error())
	}
	return c.JSON(http.StatusOK, successResponse("User profile retrieved", user))
}

func extractBearerToken(c echo.Context) string {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}
