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

	options, err := h.eventOptionUsecase.ListByEvent(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Event options retrieved", options))
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

func (h *APIHandler) DeleteEventOption(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.eventOptionUsecase.Delete(ctx, id); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "id": id}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Event option deleted", nil))
}
