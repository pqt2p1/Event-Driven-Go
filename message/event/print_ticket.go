package event

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"net/http"
	"tickets/entities"
)

func (h Handler) PrintTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	fileID := fmt.Sprintf("%s-ticket.html", event.TicketID)

	content := fmt.Sprintf(`
<html>
<body>
    <h1>Ticket: %s</h1>
    <p>Price: %s %s</p>
    <p>Customer: %s</p>
</body>
</html>
    `, event.TicketID, event.Price.Amount, event.Price.Currency, event.CustomerEmail)

	resp, err := h.filesApi.PutFilesFileIdContentWithTextBodyWithResponse(
		ctx,
		fileID,
		content,
	)
	if err != nil {
		return fmt.Errorf("failed to store ticket file: %w", err)
	}

	if resp.StatusCode() == http.StatusConflict {
		log.FromContext(ctx).With("file", fileID).Info("file already exists")
		return nil
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return nil
}
