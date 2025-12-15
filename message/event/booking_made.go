package event

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"tickets/entities"
)

func (h Handler) BookingMade(ctx context.Context, event *entities.BookingMade) error {
	// Parse UUIDs
	showID, err := uuid.Parse(event.ShowID)
	if err != nil {
		return fmt.Errorf("invalid show ID: %w", err)
	}

	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return fmt.Errorf("invalid booking ID: %w", err)
	}

	// Lấy Show để có DeadNationID
	show, err := h.showsRepo.ShowByID(ctx, showID)
	if err != nil {
		return fmt.Errorf("could not get show: %w", err)
	}

	deadNationEventID, err := uuid.Parse(show.DeadNationID)
	if err != nil {
		return fmt.Errorf("invalid dead nation event ID: %w", err)
	}

	// Gọi Dead Nation API
	err = h.deadNationClient.BookInDeadNation(
		ctx,
		entities.DeadNationBooking{
			BookingID:         bookingID,
			NumberOfTickets:   event.NumberOfTickets,
			CustomerEmail:     event.CustomerEmail,
			DeadNationEventID: deadNationEventID,
		},
	)
	if err != nil {
		return fmt.Errorf("could not call Dead Nation API: %w", err)
	}

	return nil
}
