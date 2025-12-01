package http

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"tickets/entities"
)

type ticketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
}

type TicketStatus struct {
	TicketID      string         `json:"ticket_id"`
	Status        string         `json:"status"`
	CustomerEmail string         `json:"customer_email"`
	Price         entities.Money `json:"price"`
}

type ticketsStatusRequest struct {
	Tickets []TicketStatus `json:"tickets"`
}

type TicketsStatusRequest struct {
	Tickets []TicketStatus `json:"tickets"`
}

func (h Handler) PostTicketsStatus(c echo.Context) error {
	var request ticketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	for _, ticketStatus := range request.Tickets {
		if ticketStatus.Status == "confirmed" {
			event := entities.TicketBookingConfirmed{
				Header:        entities.NewMessageHeader(),
				TicketID:      ticketStatus.TicketID,
				CustomerEmail: ticketStatus.CustomerEmail,
				Price:         ticketStatus.Price,
			}
			err = h.eventBus.Publish(c.Request().Context(), event)
			if err != nil {
				return err
			}
		} else if ticketStatus.Status == "canceled" {
			event := entities.TicketBookingCanceled{
				Header:        entities.NewMessageHeader(),
				TicketID:      ticketStatus.TicketID,
				CustomerEmail: ticketStatus.CustomerEmail,
				Price:         ticketStatus.Price,
			}
			err = h.eventBus.Publish(c.Request().Context(), event)
			if err != nil {
				return err
			}
		}
	}
	return c.NoContent(http.StatusOK)
}
