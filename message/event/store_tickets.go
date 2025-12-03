package event

import (
	"golang.org/x/net/context"
	"log/slog"
	"tickets/entities"
)

func (h Handler) StoreTickets(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	slog.Info("Storing ticket in database")

	ticket := entities.Ticket{
		TicketID:      event.TicketID,
		Price:         event.Price,
		CustomerEmail: event.CustomerEmail,
	}
	return h.ticketsRepo.Add(ctx, ticket)
}
