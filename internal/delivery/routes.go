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
		api.GET("/users/me/", h.GetMe)
		api.GET("/users/:id/", h.GetUserProfile)
		api.PATCH("/users/me/", h.UpdateMe)
		api.GET("/users/me/payment-methods/", h.ListPaymentMethods)
		api.GET("/users/me/refunds/", h.ListMyRefunds)
		api.POST("/users/me/payment-methods/", h.CreatePaymentMethod)
		api.PATCH("/users/me/payment-methods/:id/", h.UpdatePaymentMethod)
		api.DELETE("/users/me/payment-methods/:id/", h.DeletePaymentMethod)
		api.GET("/users/:id/events/created/", h.ListUserCreatedEvents)
		api.GET("/users/:id/events/participated/", h.ListUserParticipatedEvents)

		api.GET("/venues/", h.ListVenues)
		api.POST("/venues/", h.CreateVenue)
		api.PATCH("/venues/:id/", h.UpdateVenue)
		api.DELETE("/venues/:id/", h.DeleteVenue)

		api.GET("/events/", h.ListEvents)
		api.POST("/events/", h.CreateEvent)
		api.PATCH("/events/:id/status/", h.UpdateEventStatus)
		api.PATCH("/events/:id/chosen-option/", h.UpdateEventChosenOption)
		api.PATCH("/events/:id/schedule/", h.UpdateEventSchedule)
		api.PATCH("/events/:id/images/", h.UpdateEventImages)
		api.GET("/events/:id/schedule/history/", h.ListEventScheduleChangeLogs)

		api.GET("/events/:event_id/options/", h.ListEventOptions)
		api.GET("/events/:event_id/options/with-voters/", h.ListEventOptionsWithVoters)
		api.GET("/events/:event_id/options/history/", h.ListEventOptionChangeLogs)
		api.POST("/events/:event_id/options/", h.CreateEventOption)
		api.PATCH("/events/:event_id/options/:id/", h.UpdateEventOption)
		api.DELETE("/events/:event_id/options/:id/", h.DeleteEventOption)

		api.POST("/events/:event_id/votes/", h.CastVote)
		api.DELETE("/events/:event_id/votes/:option_id/", h.RemoveVote)

		api.GET("/events/:event_id/participants/", h.ListParticipants)
		api.POST("/events/:event_id/participants/", h.JoinEvent)
		api.POST("/events/share/:token/participants/", h.JoinEventByToken)
		api.DELETE("/events/:event_id/participants/", h.LeaveEvent)
		api.GET("/events/:event_id/participants/:participant_id/removal-impact/", h.PreviewRemoveParticipant)
		api.DELETE("/events/:event_id/participants/:participant_id/", h.RemoveParticipant)
		api.POST("/events/:event_id/participants/guest/", h.JoinEventAsGuest)
		api.POST("/events/share/:token/participants/guest/", h.JoinEventAsGuestByToken)

		api.GET("/events/:event_id/payment/", h.GetPayment)
		api.GET("/events/:event_id/refunds/", h.ListEventRefunds)
		api.POST("/events/:event_id/payment/", h.CreatePayment)
		api.PATCH("/events/:event_id/payment/", h.UpdatePayment)
		api.PATCH("/events/:event_id/payment/config/", h.UpdatePaymentConfig)
		api.POST("/events/:event_id/payment/claim/", h.ClaimPayment)
		api.POST("/events/:event_id/payment/charge-all/", h.ChargeAllPayments)
		api.PATCH("/events/:event_id/payment/records/:participant_id/", h.ConfirmPayment)
		api.POST("/events/:event_id/payment/records/:participant_id/adjust/", h.AdjustPayment)
		api.PATCH("/refunds/:id/destination/", h.UpdateRefundDestination)
		api.PATCH("/refunds/:id/send/", h.SendRefund)
		api.PATCH("/refunds/:id/confirm-receipt/", h.ConfirmRefundReceipt)

		api.GET("/events/:event_id/whatsapp/venue/", h.GenerateVenueWhatsAppLink)
		api.GET("/events/:event_id/whatsapp/nudge/:user_id/", h.GenerateNudgeWhatsAppLink)

		api.POST("/uploads/image/", h.UploadImage)
	}

	e.GET("/events/public/", h.ListPublicEvents)
	e.GET("/events/:token/", h.GetEventByToken)
	e.GET("/events/:token/options/", h.ListEventOptions)
	e.GET("/events/:token/options/with-voters/", h.ListEventOptionsWithVoters)
	e.GET("/events/:token/participants/", h.ListParticipants)

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, successResponse("Welcome to kumpul-be API", map[string]string{
			"message": "kumpul-be API",
		}))
	})
}
