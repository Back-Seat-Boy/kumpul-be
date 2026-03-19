package delivery

import (
	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
)

type APIHandler struct {
	authUsecase    model.AuthUsecase
	sessionUsecase model.SessionUsecase
	userUsecase    model.UserUsecase
}

func NewAPIHandler(
	authUsecase model.AuthUsecase,
	sessionUsecase model.SessionUsecase,
	userUsecase model.UserUsecase,
) *APIHandler {
	return &APIHandler{
		authUsecase:    authUsecase,
		sessionUsecase: sessionUsecase,
		userUsecase:    userUsecase,
	}
}
