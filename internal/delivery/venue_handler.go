package delivery

import (
	"net/http"
	"strconv"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) ListVenues(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse pagination parameters
	req := &model.ListVenuesRequest{
		Mode: model.PaginationModePage, // Default to page mode
	}

	// Parse mode
	if mode := c.QueryParam("mode"); mode == "cursor" {
		req.Mode = model.PaginationModeCursor
	}

	// Parse page
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}

	// Parse limit
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	// Parse cursor
	req.Cursor = c.QueryParam("cursor")

	// Parse filters
	req.Filter.Search = c.QueryParam("search")

	response, err := h.venueUsecase.ListPaginated(ctx, req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx)}).Error()
		return err
	}

	return c.JSON(http.StatusOK, successResponse("Venues list retrieved", response))
}

func (h *APIHandler) CreateVenue(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	var req model.CreateVenueRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	venue, err := h.venueUsecase.Create(ctx, user.ID, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "req": utils.Dump(req)}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Venue created", venue))
}

func (h *APIHandler) UpdateVenue(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req model.UpdateVenueRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	venue, err := h.venueUsecase.Update(ctx, id, &req)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "id": id, "req": utils.Dump(req)}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Venue updated", venue))
}

func (h *APIHandler) DeleteVenue(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.venueUsecase.Delete(ctx, id); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "id": id}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Venue deleted", nil))
}
