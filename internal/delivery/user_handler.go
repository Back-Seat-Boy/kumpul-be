package delivery

import (
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) GetUserByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	user, err := h.userUsecase.GetByID(ctx, id)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpOutGoingContext(ctx), "id": id}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("User retrieved", toUserInfo(user)))
}

func (h *APIHandler) ListUsers(c echo.Context) error {
	ctx := c.Request().Context()

	users, err := h.userUsecase.List(ctx)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpOutGoingContext(ctx)}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Users list retrieved", users))
}

func (h *APIHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var updateReq model.UpdateUserInput
	if err := c.Bind(&updateReq); err != nil {
		return err
	}

	res, err := h.userUsecase.Update(ctx, id, &updateReq)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpOutGoingContext(ctx), "id": id, "updateReq": utils.Dump(updateReq)}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("User updated", toUserInfo(res)))
}

func (h *APIHandler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.userUsecase.Delete(ctx, id); err != nil {
		log.WithFields(log.Fields{"context": utils.DumpOutGoingContext(ctx), "id": id}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("User deleted", nil))
}
