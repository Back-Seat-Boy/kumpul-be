package delivery

import (
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) ListParticipants(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")

	participants, err := h.participantUsecase.ListByEvent(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Participants retrieved", participants))
}

func (h *APIHandler) JoinEvent(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	if err := h.participantUsecase.Join(ctx, eventID, user.ID); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "userID": user.ID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Joined event", nil))
}

func (h *APIHandler) LeaveEvent(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	if err := h.participantUsecase.Leave(ctx, eventID, user.ID); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "userID": user.ID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Left event", nil))
}
