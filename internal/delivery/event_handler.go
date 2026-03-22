package delivery

import (
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) ListEvents(c echo.Context) error {
	ctx := c.Request().Context()

	events, err := h.eventUsecase.ListForDashboard(ctx)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Events list retrieved", events))
}

func (h *APIHandler) GetEventByToken(c echo.Context) error {
	ctx := c.Request().Context()
	token := c.Param("token")

	event, err := h.eventUsecase.GetByShareToken(ctx, token)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "token": token}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event retrieved", event))
}

func (h *APIHandler) CreateEvent(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	var req model.CreateEventRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	event, err := h.eventUsecase.Create(ctx, user.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event created", event))
}

func (h *APIHandler) UpdateEventStatus(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req model.UpdateEventStatusRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.eventUsecase.UpdateStatus(ctx, id, req.Status); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "id": id, "status": req.Status}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event status updated", nil))
}

func (h *APIHandler) UpdateEventChosenOption(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req model.UpdateEventChosenOptionRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.eventUsecase.UpdateChosenOption(ctx, id, req.OptionID); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "id": id, "optionID": req.OptionID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event chosen option updated", nil))
}
