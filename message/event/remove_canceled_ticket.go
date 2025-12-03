package event

import (
	"context"
	"log/slog"
	"tickets/entities"
)

func (h Handler) RemoveCanceledTicket(ctx context.Context, event *entities.TicketBookingCanceled) error {
	slog.Info("Removing canceled ticket from database")

	return h.ticketsRepo.Remove(ctx, event.TicketID)
}
