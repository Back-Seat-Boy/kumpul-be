package delivery

import (
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) GetPayment(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")

	payment, err := h.paymentUsecase.GetByEventID(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return err
	}

	recordsWithSummary, err := h.paymentRecordUsecase.GetByPaymentID(ctx, payment.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "paymentID": payment.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment retrieved", map[string]interface{}{
		"payment": payment,
		"records": recordsWithSummary.Records,
		"summary": map[string]interface{}{
			"num_participants":     recordsWithSummary.NumParticipants,
			"num_confirmed":        recordsWithSummary.NumConfirmed,
			"num_claimed":          recordsWithSummary.NumClaimed,
			"num_pending":          recordsWithSummary.NumPending,
			"total_collected":      recordsWithSummary.TotalCollected,
			"total_should_collect": recordsWithSummary.TotalShouldCollect,
			"balance":              recordsWithSummary.Balance,
			"per_person_status":    recordsWithSummary.PerPersonStatus,
		},
	}))
}

func (h *APIHandler) CreatePayment(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")

	var req model.CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	payment, err := h.paymentUsecase.Create(ctx, eventID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment created", payment))
}

func (h *APIHandler) ClaimPayment(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	var req model.ClaimPaymentRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	payment, err := h.paymentUsecase.GetByEventID(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return err
	}

	if err := h.paymentRecordUsecase.Claim(ctx, payment.ID, user.ID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "paymentID": payment.ID, "userID": user.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment claimed", nil))
}

func (h *APIHandler) ConfirmPayment(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")
	userID := c.Param("user_id")

	payment, err := h.paymentUsecase.GetByEventID(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return err
	}

	if err := h.paymentRecordUsecase.Confirm(ctx, payment.ID, userID); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "paymentID": payment.ID, "userID": userID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment confirmed", nil))
}

func (h *APIHandler) AdjustPayment(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")
	userID := c.Param("user_id")

	var req model.AdjustPaymentRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	payment, err := h.paymentUsecase.GetByEventID(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return err
	}

	if err := h.paymentRecordUsecase.AdjustPayment(ctx, payment.ID, userID, requester.ID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "paymentID": payment.ID, "userID": userID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment adjusted", nil))
}
