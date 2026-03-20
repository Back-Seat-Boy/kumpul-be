package delivery

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) GenerateVenueWhatsAppLink(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")

	event, err := h.eventUsecase.GetByID(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if event.ChosenOptionID == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "event has no chosen option")
	}

	option, err := h.eventOptionUsecase.GetByID(ctx, *event.ChosenOptionID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "optionID": *event.ChosenOptionID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	venue, err := h.venueUsecase.GetByID(ctx, option.VenueID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "venueID": option.VenueID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	message := fmt.Sprintf("Halo, kami ingin memesan lapangan untuk %s pada %s pukul %s-%s. Apakah tersedia? Terima kasih.",
		event.Title,
		option.Date.Format("2006-01-02"),
		option.StartTime,
		option.EndTime,
	)

	waLink := fmt.Sprintf("https://wa.me/%s?text=%s", venue.WhatsappNumber, url.QueryEscape(message))

	return c.JSON(http.StatusOK, successResponse("WhatsApp link generated", map[string]string{
		"link": waLink,
	}))
}

func (h *APIHandler) GenerateNudgeWhatsAppLink(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := c.Param("event_id")
	userID := c.Param("user_id")

	event, err := h.eventUsecase.GetByID(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	participant, err := h.userUsecase.GetByID(ctx, userID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "userID": userID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	payment, err := h.paymentUsecase.GetByEventID(ctx, eventID)
	if err != nil {
		log.WithFields(log.Fields{"context": utils.DumpIncomingContext(ctx), "eventID": eventID}).Error()
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	message := fmt.Sprintf("Hei %s, jangan lupa transfer Rp%d untuk %s ke %s. Terima kasih ya!",
		participant.Name,
		payment.SplitAmount,
		event.Title,
		payment.PaymentInfo,
	)

	waLink := fmt.Sprintf("https://wa.me/%s?text=%s", participant.WhatsappNumber, url.QueryEscape(message))

	return c.JSON(http.StatusOK, successResponse("Nudge link generated", map[string]string{
		"link": waLink,
	}))
}
