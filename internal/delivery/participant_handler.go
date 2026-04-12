package delivery

import (
	"net/http"
	"strconv"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) ListParticipants(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")
	req := parseListParticipantsRequest(c)

	// If accessed via public route with token, look up event by share token
	if eventID == "" {
		token := c.Param("token")
		event, err := h.eventUsecase.GetByShareToken(ctx, token)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "event not found")
		}
		eventID = event.ID
	}
	req.EventID = eventID

	participants, err := h.participantUsecase.ListByEvent(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Participants retrieved", participants))
}

func parseListParticipantsRequest(c echo.Context) *model.ListParticipantsRequest {
	req := &model.ListParticipantsRequest{
		Mode:      model.PaginationModePage,
		SortOrder: model.ParticipantSortOrderAsc,
	}

	if mode := c.QueryParam("mode"); mode == "cursor" {
		req.Mode = model.PaginationModeCursor
	}
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	req.Cursor = c.QueryParam("cursor")
	req.Filter.Search = c.QueryParam("search")
	if sortOrder := c.QueryParam("sort_order"); sortOrder != "" {
		req.SortOrder = model.ParticipantSortOrder(sortOrder)
	}

	return req
}

func (h *APIHandler) JoinEvent(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	if err := h.participantUsecase.Join(ctx, eventID, user.ID, false); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "userID": user.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Joined event", nil))
}

func (h *APIHandler) JoinEventByToken(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	token := c.Param("token")

	event, err := h.eventUsecase.GetByShareToken(ctx, token)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "token": token, "userID": user.ID}).Error()
		return err
	}

	if err := h.participantUsecase.Join(ctx, event.ID, user.ID, true); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": event.ID, "token": token, "userID": user.ID}).Error()
		return err
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

	if err := h.participantUsecase.JoinAsGuest(ctx, user.ID, eventID, &req, false); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "userID": user.ID, "guestName": req.GuestName}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Joined event", nil))
}

func (h *APIHandler) JoinEventAsGuestByToken(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	token := c.Param("token")

	var req model.JoinAsGuestRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	event, err := h.eventUsecase.GetByShareToken(ctx, token)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "token": token, "userID": user.ID, "guestName": req.GuestName}).Error()
		return err
	}

	if err := h.participantUsecase.JoinAsGuest(ctx, user.ID, event.ID, &req, true); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": event.ID, "token": token, "userID": user.ID, "guestName": req.GuestName}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Joined event", nil))
}
