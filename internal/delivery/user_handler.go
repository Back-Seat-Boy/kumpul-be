package delivery

import (
	"net/http"

	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) UpdateMe(c echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get(string(model.ContextKeyUser)).(UserInfo)

	var updateReq model.UpdateUserInput
	if err := c.Bind(&updateReq); err != nil {
		return err
	}
	if err := c.Validate(&updateReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	res, err := h.userUsecase.Update(ctx, user.ID, &updateReq)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "userID": user.ID, "updateReq": utils.Dump(updateReq)}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, successResponse("Profile updated", toUserInfo(res)))
}
