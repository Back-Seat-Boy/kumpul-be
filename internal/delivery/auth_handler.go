package delivery

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Back-Seat-Boy/kumpul-be/internal/config"
	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) GoogleLogin(c echo.Context) error {
	ctx := c.Request().Context()
	url := h.authUsecase.GetGoogleLoginURL(ctx)
	return c.JSON(http.StatusOK, successResponse("Google login URL generated", LoginURLResponse{LoginURL: url}))
}

func (h *APIHandler) GoogleCallback(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")

	session, user, err := h.authUsecase.HandleGoogleCallback(ctx, code)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "code": code}).Error()
		// Redirect to frontend with error
		redirectURL := fmt.Sprintf("%s/auth/callback?error=%s", config.FrontendURL(), err.Error())
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Redirect to frontend with session data including expires_at
	redirectURL := fmt.Sprintf("%s/auth/callback?session_id=%s&expires_at=%s&user_id=%s&user_name=%s&user_email=%s&email_verified=%t&avatar_url=%s",
		config.FrontendURL(),
		session.ID,
		session.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		user.ID,
		urlEncode(user.Name),
		urlEncode(user.Email),
		user.EmailVerified,
		urlEncode(user.AvatarURL),
	)
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func urlEncode(s string) string {
	return strings.ReplaceAll(s, " ", "%20")
}

func (h *APIHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	sessionID := extractBearerToken(c)

	if sessionID != "" {
		err := h.authUsecase.Logout(ctx, sessionID)
		if err != nil {
			log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "sessionID": sessionID}).Error()
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
