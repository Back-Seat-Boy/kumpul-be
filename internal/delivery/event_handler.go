package delivery

import (
	"net/http"
	"strconv"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) ListEvents(c echo.Context) error {
	ctx := c.Request().Context()
	req := parseListEventsRequest(c)
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)
	req.RequesterUserID = user.ID

	response, err := h.eventUsecase.ListForDashboard(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Events list retrieved", response))
}

func (h *APIHandler) ListPublicEvents(c echo.Context) error {
	ctx := c.Request().Context()
	req := parseListEventsRequest(c)
	req.Filter.Visibility = model.EventVisibilityPublic

	response, err := h.eventUsecase.ListPublic(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Public events list retrieved", response))
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

func (h *APIHandler) UpdateEventSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	var req model.UpdateEventScheduleRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.eventUsecase.UpdateSchedule(ctx, id, user.ID, &req); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "id": id, "req": utils.Dump(req)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event schedule updated", nil))
}

func (h *APIHandler) ListEventScheduleChangeLogs(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	logs, err := h.eventUsecase.ListScheduleChangeLogs(ctx, id, user.ID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "id": id}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Event schedule change logs retrieved", logs))
}

func parseListEventsRequest(c echo.Context) *model.ListEventsRequest {
	req := &model.ListEventsRequest{
		Mode: model.PaginationModePage,
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
	req.Filter.Status = model.EventStatus(c.QueryParam("status"))
	req.Filter.Visibility = model.EventVisibility(c.QueryParam("visibility"))
	req.Filter.EventDate = c.QueryParam("event_date")

	return req
}
