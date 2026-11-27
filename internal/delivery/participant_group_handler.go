package delivery

import (
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) ListParticipantGroups(c echo.Context) error {
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

	groups, err := h.participantGroupUsecase.List(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Participant groups retrieved", groups))
}

func (h *APIHandler) CreateParticipantGroup(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	var req model.CreateParticipantGroupRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	group, err := h.participantGroupUsecase.Create(ctx, eventID, requester.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "requesterID": requester.ID, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Participant group created", group))
}

func (h *APIHandler) UpdateParticipantGroup(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")
	groupID := c.Param("group_id")

	var req model.UpdateParticipantGroupRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	group, err := h.participantGroupUsecase.Update(ctx, eventID, groupID, requester.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "groupID": groupID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Participant group updated", group))
}

func (h *APIHandler) DeleteParticipantGroup(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")
	groupID := c.Param("group_id")

	if err := h.participantGroupUsecase.Delete(ctx, eventID, groupID, requester.ID); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "groupID": groupID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Participant group deleted", nil))
}

func (h *APIHandler) AssignParticipantGroup(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")
	participantID := c.Param("participant_id")

	var req model.AssignParticipantGroupRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.participantGroupUsecase.AssignParticipant(ctx, eventID, participantID, requester.ID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "participantID": participantID, "requesterID": requester.ID}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Participant group assigned", nil))
}

func (h *APIHandler) UpdateParticipantGroupSettings(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	var req model.UpdateParticipantGroupSettingsRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.participantGroupUsecase.UpdateSettings(ctx, eventID, requester.ID, req.AllowGroupSelfSelect); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "requesterID": requester.ID, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Participant group settings updated", nil))
}

func (h *APIHandler) ChangeOwnParticipantGroup(c echo.Context) error {
	ctx := c.Request().Context()
	requester := c.Get(string(model.ContextKeyUser)).(UserInfo)
	eventID := c.Param("event_id")

	var req model.AssignParticipantGroupRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.participantGroupUsecase.ChangeOwnGroup(ctx, eventID, requester.ID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID, "requesterID": requester.ID, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Participant group changed", nil))
}
