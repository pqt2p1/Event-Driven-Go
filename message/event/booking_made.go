package event

import (
	"context"
	"fmt"
	"tickets/entities"
)

func (h Handler) BookingMade(ctx context.Context, event *entities.BookingMade) error {
	// Lấy Show để có DeadNationID
	show, err := h.showsRepo.ShowByID(ctx, event.ShowID)
	if err != nil {
		return fmt.Errorf("could not get show: %w", err)
	}

	// Gọi Dead Nation API
	err = h.deadNationClient.PostTicketBooking(
		ctx,
		event.BookingID,
		show.DeadNationID,
		event.NumberOfTickets,
		event.CustomerEmail,
	)
	if err != nil {
		return fmt.Errorf("could not call Dead Nation API: %w", err)
	}

	return nil
}
