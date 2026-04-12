package delivery

import (
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) ListEventRefunds(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	refunds, err := h.refundUsecase.ListByEvent(ctx, eventID, requester.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Refunds retrieved", refunds))
}

func (h *APIHandler) ListMyRefunds(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	refunds, err := h.refundUsecase.ListByUserID(ctx, user.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "userID": user.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Refunds retrieved", refunds))
}

func (h *APIHandler) UpdateRefundDestination(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	id := c.Param("id")

	var req model.UpdateRefundDestinationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	refund, err := h.refundUsecase.UpdateDestination(ctx, id, user.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "refundID": id, "userID": user.ID, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Refund destination updated", refund))
}

func (h *APIHandler) SendRefund(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	id := c.Param("id")

	var req model.SendRefundRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	refund, err := h.refundUsecase.MarkSent(ctx, id, requester.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "refundID": id, "requesterID": requester.ID, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Refund marked as sent", refund))
}

func (h *APIHandler) ConfirmRefundReceipt(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	id := c.Param("id")

	refund, err := h.refundUsecase.ConfirmReceipt(ctx, id, user.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "refundID": id, "userID": user.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Refund receipt confirmed", refund))
}
