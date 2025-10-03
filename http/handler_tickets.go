package http

import (
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
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

func (h Handler) PostTicketsStatus(c echo.Context) error {
	var request ticketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	for _, ticketStatus := range request.Tickets {
		if ticketStatus.Status == "confirmed" {
			payload := entities.TicketBookingConfirmed{
				Header:        entities.NewMessageHeader(),
				TicketID:      ticketStatus.TicketID,
				CustomerEmail: ticketStatus.CustomerEmail,
				Price:         ticketStatus.Price,
			}
			payloadJSON, err := json.Marshal(payload)
			msg := message.NewMessage(watermill.NewUUID(), []byte(payloadJSON))
			msg.Metadata.Set("correlation_id", c.Request().Header.Get("Correlation-ID"))
			err = h.publisher.Publish("TicketBookingConfirmed", msg)
			if err != nil {
				return err
			}
		} else if ticketStatus.Status == "canceled" {
			payload := entities.TicketBookingCanceled{
				Header:        entities.NewMessageHeader(),
				TicketID:      ticketStatus.TicketID,
				CustomerEmail: ticketStatus.CustomerEmail,
				Price:         ticketStatus.Price,
			}
			payloadJSON, err := json.Marshal(payload)
			msg := message.NewMessage(watermill.NewUUID(), []byte(payloadJSON))
			msg.Metadata.Set("correlation_id", c.Request().Header.Get("Correlation-ID"))
			err = h.publisher.Publish("TicketBookingCanceled", msg)
			if err != nil {
				return err
			}
		}
	}
	return c.NoContent(http.StatusOK)
}
