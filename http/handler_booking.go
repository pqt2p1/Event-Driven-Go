package http

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
)

func (h *Handler) PostBookTickets(c echo.Context) error {
	var request = entities.Booking{}
	err := c.Bind(&request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if request.NumberOfTickets < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "number of tickets must be greater than 0")
	}

	bookingId := request.BookingID
	if bookingId == "" {
		bookingId = uuid.New().String()
	}

	request.BookingID = bookingId

	err = h.bookingRepo.Add(c.Request().Context(), request)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"booking_id": bookingId,
	})

}
