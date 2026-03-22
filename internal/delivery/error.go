package delivery

import (
	"errors"
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/labstack/echo/v4"
)

// ErrorResponse represents the standard error response
type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// CustomHTTPErrorHandler handles errors and returns consistent format
func CustomHTTPErrorHandler(err error, c echo.Context) {
	// Get the status code
	code := http.StatusInternalServerError
	message := "Internal server error"

	// Check if it's an Echo HTTPError
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		code = httpErr.Code
		if msg, ok := httpErr.Message.(string); ok {
			message = msg
		} else if msg, ok := httpErr.Message.(error); ok {
			message = msg.Error()
		}
	}

	// Check if it's an AppError
	var appErr *model.AppError
	if errors.As(err, &appErr) {
		code = appErr.StatusCode
		message = appErr.Message
		if message == "" && appErr.Err != nil {
			message = appErr.Err.Error()
		}
	}

	// Handle specific known errors
	switch {
	case errors.Is(err, model.ErrUserNotFound) || errors.Is(err, model.ErrSessionNotFound) ||
		errors.Is(err, model.ErrEventNotFound) || errors.Is(err, model.ErrEventOptionNotFound) ||
		errors.Is(err, model.ErrPaymentNotFound) || errors.Is(err, model.ErrPaymentRecordNotFound) ||
		errors.Is(err, model.ErrVenueNotFound) || errors.Is(err, model.ErrParticipantNotFound) ||
		errors.Is(err, model.ErrVoteNotFound):
		code = http.StatusNotFound
		message = "Resource not found"
	case errors.Is(err, model.ErrUnauthorized) || errors.Is(err, model.ErrInvalidSession):
		code = http.StatusUnauthorized
		message = "Unauthorized"
	case errors.Is(err, model.ErrForbidden):
		code = http.StatusForbidden
		message = "Forbidden"
	case errors.Is(err, model.ErrInvalidEmail) || errors.Is(err, model.ErrInvalidGoogleID) ||
		errors.Is(err, model.ErrNoParticipantsInEvent) || errors.Is(err, model.ErrAlreadyVoted) ||
		errors.Is(err, model.ErrNotVoted):
		code = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, model.ErrDuplicateEmail) || errors.Is(err, model.ErrDuplicateGoogleID):
		code = http.StatusConflict
		message = "Resource already exists"
	}

	// Don't override message if already set from HTTPError or AppError
	if errors.As(err, &httpErr) || errors.As(err, &appErr) {
		// message already set above
	} else if code == http.StatusInternalServerError {
		// For unknown internal errors, show generic message in production
		message = err.Error()
	}

	response := ErrorResponse{
		Status:  "ERROR",
		Message: message,
		Data:    nil,
	}

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, response)
		}
		if err != nil {
			c.Logger().Error(err)
		}
	}
}
