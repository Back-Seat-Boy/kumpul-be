package delivery

import (
	"errors"
	"io"
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) GetPayment(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")
	statusFilter, err := parsePaymentRecordStatus(c.QueryParam("status"))
	if err != nil {
		return err
	}

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

	splitBill, err := h.paymentUsecase.GetSplitBillDetails(ctx, payment.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "paymentID": payment.ID}).Error()
		return err
	}

	records := recordsWithSummary.Records
	perPersonStatus := recordsWithSummary.PerPersonStatus
	if statusFilter != "" {
		records = filterPaymentRecordsByStatus(records, statusFilter)
		perPersonStatus = filterParticipantPaymentStatusByStatus(perPersonStatus, statusFilter)
	}

	return c.JSON(http.StatusOK, successResponse("Payment retrieved", map[string]interface{}{
		"payment":    payment,
		"records":    records,
		"split_bill": splitBill,
		"summary": map[string]interface{}{
			"num_participants":     recordsWithSummary.NumParticipants,
			"num_confirmed":        recordsWithSummary.NumConfirmed,
			"num_claimed":          recordsWithSummary.NumClaimed,
			"num_pending":          recordsWithSummary.NumPending,
			"total_collected":      recordsWithSummary.TotalCollected,
			"total_should_collect": recordsWithSummary.TotalShouldCollect,
			"balance":              recordsWithSummary.Balance,
			"per_person_status":    perPersonStatus,
		},
	}))
}

func parsePaymentRecordStatus(raw string) (model.PaymentRecordStatus, error) {
	status := model.PaymentRecordStatus(raw)
	switch status {
	case "":
		return "", nil
	case model.PaymentRecordStatusPending, model.PaymentRecordStatusClaimed, model.PaymentRecordStatusConfirmed:
		return status, nil
	default:
		return "", echo.NewHTTPError(http.StatusBadRequest, "status must be one of: pending, claimed, confirmed")
	}
}

func filterPaymentRecordsByStatus(records []*model.PaymentRecord, status model.PaymentRecordStatus) []*model.PaymentRecord {
	filtered := make([]*model.PaymentRecord, 0, len(records))
	for _, record := range records {
		if record.Status == status {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func filterParticipantPaymentStatusByStatus(statuses []model.ParticipantPaymentStatus, status model.PaymentRecordStatus) []model.ParticipantPaymentStatus {
	filtered := make([]model.ParticipantPaymentStatus, 0, len(statuses))
	for _, participantStatus := range statuses {
		if participantStatus.Status == string(status) {
			filtered = append(filtered, participantStatus)
		}
	}
	return filtered
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
	if req.PaymentMethodID == "" && req.PaymentInfo == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "payment_info or payment_method_id is required")
	}
	switch req.Type {
	case "", string(model.PaymentTypeTotal):
		if req.TotalCost <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "total_cost must be greater than 0 for total payment type")
		}
	case string(model.PaymentTypePerPerson):
		if req.PerPersonAmount <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "per_person_amount must be greater than 0 for per_person payment type")
		}
	case string(model.PaymentTypeSplitBill):
		if req.TaxAmount < 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "tax_amount must be greater than or equal to 0")
		}
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "type must be either total, per_person, or split_bill")
	}

	payment, err := h.paymentUsecase.Create(ctx, eventID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment created", payment))
}

func (h *APIHandler) UpdatePayment(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	var req model.UpdatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if req.PaymentMethodID == "" && req.PaymentInfo == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "payment_info or payment_method_id is required")
	}

	payment, err := h.paymentUsecase.UpdatePaymentInfo(ctx, eventID, requester.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment updated", payment))
}

func (h *APIHandler) UpdatePaymentConfig(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	var req model.UpdatePaymentConfigRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	switch req.Type {
	case "", string(model.PaymentTypeTotal):
		if req.TotalCost <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "total_cost must be greater than 0 for total payment type")
		}
	case string(model.PaymentTypePerPerson):
		if req.PerPersonAmount <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "per_person_amount must be greater than 0 for per_person payment type")
		}
	case string(model.PaymentTypeSplitBill):
		if req.TaxAmount < 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "tax_amount must be greater than or equal to 0")
		}
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "type must be either total, per_person, or split_bill")
	}

	payment, err := h.paymentUsecase.UpdatePaymentConfig(ctx, eventID, requester.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "requesterID": requester.ID, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment configuration updated", payment))
}

func (h *APIHandler) ClaimPayment(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	var req model.ClaimPaymentRequest
	if err := c.Bind(&req); err != nil && !errors.Is(err, io.EOF) {
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
	participantID := c.Param("participant_id")

	var req model.ConfirmPaymentRequest
	if err := c.Bind(&req); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if req.Amount != nil && *req.Amount < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "amount must be greater than or equal to 0")
	}

	payment, err := h.paymentUsecase.GetByEventID(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return err
	}

	if err := h.paymentRecordUsecase.Confirm(ctx, payment.ID, participantID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "paymentID": payment.ID, "participantID": participantID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment confirmed", nil))
}

func (h *APIHandler) AdjustPayment(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")
	participantID := c.Param("participant_id")

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

	if err := h.paymentRecordUsecase.AdjustPayment(ctx, payment.ID, participantID, requester.ID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "paymentID": payment.ID, "participantID": participantID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment adjusted", nil))
}

func (h *APIHandler) ChargeAllPayments(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	var req model.ChargeAllRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.paymentUsecase.ChargeAll(ctx, eventID, requester.ID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Charge applied to all payment records", nil))
}
