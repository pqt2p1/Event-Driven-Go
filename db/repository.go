package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"tickets/entities"
)

type TicketsRepository struct {
	db *sqlx.DB
}

func NewTicketsRepository(db *sqlx.DB) *TicketsRepository {
	return &TicketsRepository{db: db}
}

func (r *TicketsRepository) Add(ctx context.Context, ticket entities.Ticket) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO tickets (ticket_id, price_amount, price_currency, customer_email)
				VALUES ($1, $2, $3, $4)`,
		ticket.TicketID,
		ticket.Price.Amount,
		ticket.Price.Currency,
		ticket.CustomerEmail,
	)
	return err
}

func (r *TicketsRepository) Remove(ctx context.Context, ticketID string) error {
	_, err := r.db.ExecContext(
		ctx, `DELETE FROM tickets WHERE ticket_id = $1`,
		ticketID,
	)
	return err
}

func (r *TicketsRepository) FindAll(ctx context.Context) ([]entities.Ticket, error) {
	var tickets []entities.Ticket

	err := r.db.SelectContext(
		ctx, &tickets,
		`SELECT ticket_id, 
       price_amount as "price.amount",
       price_currency as "price.currency",
       customer_email
       FROM tickets`,
	)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}
