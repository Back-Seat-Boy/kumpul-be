package delivery

import (
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) ListPaymentMethods(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	paymentMethods, err := h.paymentMethodUsecase.ListByUserID(ctx, user.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "userID": user.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment methods retrieved", paymentMethods))
}

func (h *APIHandler) CreatePaymentMethod(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	var req model.CreatePaymentMethodRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	paymentMethod, err := h.paymentMethodUsecase.Create(ctx, user.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "userID": user.ID, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment method created", paymentMethod))
}

func (h *APIHandler) UpdatePaymentMethod(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	id := c.Param("id")

	var req model.UpdatePaymentMethodRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	paymentMethod, err := h.paymentMethodUsecase.Update(ctx, id, user.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "userID": user.ID, "paymentMethodID": id, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment method updated", paymentMethod))
}

func (h *APIHandler) DeletePaymentMethod(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	id := c.Param("id")

	if err := h.paymentMethodUsecase.Delete(ctx, id, user.ID); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "userID": user.ID, "paymentMethodID": id}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Payment method deleted", nil))
}
