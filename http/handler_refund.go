package http

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
)

func (h Handler) PutTicketRefund(c echo.Context) error {
	ticketID := c.Param("ticket_id")

	command := entities.RefundTicket{
		Header:   entities.NewMessageHeader(),
		TicketID: ticketID,
	}

	err := h.commandBus.Send(c.Request().Context(), command)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusAccepted)
}
