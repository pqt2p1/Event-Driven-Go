package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

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
		sql.LevelSerializable,
		func(ctx context.Context, tx *sqlx.Tx) error {
			// Get available seats from shows table
			availableSeats := 0
			err := tx.GetContext(ctx, &availableSeats, `
				SELECT
					number_of_tickets AS available_seats
				FROM
					shows
				WHERE
					show_id = $1
			`, booking.ShowID)
			if err != nil {
				return fmt.Errorf("could not get available seats: %w", err)
			}

			// Get already booked seats
			alreadyBookedSeats := 0
			err = tx.GetContext(ctx, &alreadyBookedSeats, `
				SELECT
					COALESCE(SUM(number_of_tickets), 0) AS already_booked_seats
				FROM
					bookings
				WHERE
					show_id = $1
			`, booking.ShowID)
			if err != nil {
				return fmt.Errorf("could not get already booked seats: %w", err)
			}

			// Check if enough seats available
			if availableSeats-alreadyBookedSeats < booking.NumberOfTickets {
				return echo.NewHTTPError(http.StatusBadRequest, "not enough seats available")
			}

			// Insert booking
			_, err = tx.NamedExecContext(ctx, `
				INSERT INTO
					bookings (booking_id, show_id, number_of_tickets, customer_email)
				VALUES (:booking_id, :show_id, :number_of_tickets, :customer_email)
			`, booking)
			if err != nil {
				return fmt.Errorf("could not insert booking: %w", err)
			}

			// Publish event
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
