package delivery

import (
	"net/http"
	"time"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/labstack/echo/v4"
)

type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type UserInfo struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	EmailVerified  bool      `json:"email_verified"`
	WhatsappNumber string    `json:"whatsapp_number,omitempty"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type AuthResponse struct {
	SessionID string   `json:"session_id"`
	User      UserInfo `json:"user"`
}

type LoginURLResponse struct {
	LoginURL string `json:"login_url"`
}

func successResponse(message string, data interface{}) SuccessResponse {
	return SuccessResponse{
		Status:  "SUCCESS",
		Message: message,
		Data:    data,
	}
}

func toUserInfo(u *model.User) UserInfo {
	return UserInfo{
		ID:             u.ID,
		Name:           u.Name,
		Email:          u.Email,
		EmailVerified:  u.EmailVerified,
		WhatsappNumber: u.WhatsappNumber,
		AvatarURL:      u.AvatarURL,
		CreatedAt:      u.CreatedAt,
	}
}

func (h *APIHandler) PingHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, successResponse("Ping successful", "Pong!"))
}
