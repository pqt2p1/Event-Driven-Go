package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"tickets/entities"
	"tickets/message/outbox"
)

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (b *BookingRepository) Add(ctx context.Context, booking entities.Booking) error {
	return updateInTx(
		ctx,
		b.db,
		sql.LevelRepeatableRead,
		func(ctx context.Context, tx *sqlx.Tx) error {
			_, err := tx.NamedExecContext(ctx, `
                                INSERT INTO 
                                        bookings (booking_id, show_id, number_of_tickets, customer_email) 
                                VALUES (:booking_id, :show_id, :number_of_tickets, :customer_email)
                        `, booking)
			if err != nil {
				return fmt.Errorf("could not insert booking: %w", err)
			}

			err = outbox.PublishEventInTx(ctx, tx, &entities.BookingMade{
				Header:          entities.NewMessageHeader(),
				BookingID:       booking.BookingID,
				NumberOfTickets: booking.NumberOfTickets,
				CustomerEmail:   booking.CustomerEmail,
				ShowID:          booking.ShowID,
			})
			if err != nil {
				return fmt.Errorf("could not publish booking made event: %w", err)
			}

			return nil
		},
	)
}
