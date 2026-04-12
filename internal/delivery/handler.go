package delivery

import (
	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
)

type APIHandler struct {
	authUsecase          model.AuthUsecase
	sessionUsecase       model.SessionUsecase
	userUsecase          model.UserUsecase
	venueUsecase         model.VenueUsecase
	eventUsecase         model.EventUsecase
	eventOptionUsecase   model.EventOptionUsecase
	voteUsecase          model.VoteUsecase
	participantUsecase   model.ParticipantUsecase
	paymentMethodUsecase model.PaymentMethodUsecase
	paymentUsecase       model.PaymentUsecase
	paymentRecordUsecase model.PaymentRecordUsecase
	refundUsecase        model.RefundUsecase
	uploadUsecase        model.UploadUsecase
}

func NewAPIHandler(
	authUsecase model.AuthUsecase,
	sessionUsecase model.SessionUsecase,
	userUsecase model.UserUsecase,
	venueUsecase model.VenueUsecase,
	eventUsecase model.EventUsecase,
	eventOptionUsecase model.EventOptionUsecase,
	voteUsecase model.VoteUsecase,
	participantUsecase model.ParticipantUsecase,
	paymentMethodUsecase model.PaymentMethodUsecase,
	paymentUsecase model.PaymentUsecase,
	paymentRecordUsecase model.PaymentRecordUsecase,
	refundUsecase model.RefundUsecase,
	uploadUsecase model.UploadUsecase,
) *APIHandler {
	return &APIHandler{
		authUsecase:          authUsecase,
		sessionUsecase:       sessionUsecase,
		userUsecase:          userUsecase,
		venueUsecase:         venueUsecase,
		eventUsecase:         eventUsecase,
		eventOptionUsecase:   eventOptionUsecase,
		voteUsecase:          voteUsecase,
		participantUsecase:   participantUsecase,
		paymentMethodUsecase: paymentMethodUsecase,
		paymentUsecase:       paymentUsecase,
		paymentRecordUsecase: paymentRecordUsecase,
		refundUsecase:        refundUsecase,
		uploadUsecase:        uploadUsecase,
	}
}
