package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"tickets/entities"
)

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (b *BookingRepository) Add(ctx context.Context, booking entities.Booking) error {
	_, err := b.db.ExecContext(
		ctx,
		`INSERT INTO bookings (booking_id, show_id, number_of_tickets, customer_email)
				VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
		booking.BookingID,
		booking.ShowID,
		booking.NumberOfTickets,
		booking.CustomerEmail,
	)

	return err
}
