package http

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h Handler) GetTickets(c echo.Context) error {
	ctx := c.Request().Context()

	tickets, err := h.ticketsRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tickets)
}
