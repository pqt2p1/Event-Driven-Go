package adapters

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/dead_nation"
	"github.com/google/uuid"
)

type DeadNationClient struct {
	clients *clients.Clients
}

func NewDeadNationClient(clients *clients.Clients) *DeadNationClient {
	return &DeadNationClient{clients: clients}
}

func (c *DeadNationClient) PostTicketBooking(
	ctx context.Context,
	bookingID string,
	eventID string,
	numberOfTickets int,
	customerEmail string,
) error {
	// Parse UUIDs
	bookingUUID, err := uuid.Parse(bookingID)
	if err != nil {
		return fmt.Errorf("invalid booking ID: %w", err)
	}

	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return fmt.Errorf("invalid event ID: %w", err)
	}

	resp, err := c.clients.DeadNation.PostTicketBookingWithResponse(
		ctx,
		dead_nation.PostTicketBookingRequest{
			BookingId:       bookingUUID,
			EventId:         eventUUID,
			NumberOfTickets: numberOfTickets,
			CustomerAddress: customerEmail,
		},
	)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("Dead Nation API returned %d", resp.StatusCode())
	}

	return nil
}
