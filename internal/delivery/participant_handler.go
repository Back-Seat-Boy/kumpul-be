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

	// If accessed via public route with token, look up event by share token
	if eventID == "" {
		token := c.Param("token")
		event, err := h.eventUsecase.GetByShareToken(ctx, token)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "event not found")
		}
		eventID = event.ID
	}

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

func (h *APIHandler) RemoveParticipant(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")
	participantID := c.Param("participant_id")

	result, err := h.participantUsecase.RemoveParticipant(ctx, eventID, participantID, requester.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "participantID": participantID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Participant removed", result))
}

func (h *APIHandler) PreviewRemoveParticipant(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")
	participantID := c.Param("participant_id")

	result, err := h.participantUsecase.PreviewRemoveParticipant(ctx, eventID, participantID, requester.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "participantID": participantID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Participant removal impact retrieved", result))
}

func (h *APIHandler) JoinEventAsGuest(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	var req model.JoinAsGuestRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.participantUsecase.JoinAsGuest(ctx, user.ID, eventID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "userID": user.ID, "guestName": req.GuestName}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Joined event", nil))
}
