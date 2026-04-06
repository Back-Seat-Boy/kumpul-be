package delivery

import (
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) ListEventOptions(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")

	// If accessed via public route with token, look up event by share token
	if eventID == "" {
		token := c.Param("token")
		event, err := h.eventUsecase.GetByShareToken(ctx, token)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "event not found")
		}
		eventID = event.ID
	}

	// Get userID from context if authenticated
	var userID *string
	if user, ok := c.Get(string(model.ContextKeyUser)).(UserInfo); ok {
		userID = &user.ID
	}

	options, err := h.eventOptionUsecase.ListByEvent(ctx, eventID, userID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event options retrieved", options))
}

func (h *APIHandler) ListEventOptionChangeLogs(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	logs, err := h.eventOptionUsecase.ListChangeLogs(ctx, eventID, user.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event option change logs retrieved", logs))
}

func (h *APIHandler) CreateEventOption(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")

	var req model.CreateEventOptionRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	option, err := h.eventOptionUsecase.Create(ctx, eventID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "req": utils.Dump(req)}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Event option created", option))
}

func (h *APIHandler) UpdateEventOption(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")
	optionID := c.Param("id")
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	var req model.UpdateEventOptionRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.eventOptionUsecase.Update(ctx, eventID, optionID, user.ID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "optionID": optionID, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event option updated", nil))
}

func (h *APIHandler) DeleteEventOption(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.eventOptionUsecase.Delete(ctx, id); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "id": id}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Event option deleted", nil))
}

func (h *APIHandler) ListEventOptionsWithVoters(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")

	// If accessed via public route with token, look up event by share token
	if eventID == "" {
		token := c.Param("token")
		event, err := h.eventUsecase.GetByShareToken(ctx, token)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "event not found")
		}
		eventID = event.ID
	}

	// Get userID from context if authenticated
	var userID *string
	if user, ok := c.Get(string(model.ContextKeyUser)).(UserInfo); ok {
		userID = &user.ID
	}

	options, err := h.eventOptionUsecase.ListByEventWithVoters(ctx, eventID, userID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event options with voters retrieved", options))
}
