package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"tickets/entities"
)

type ShowsRepository struct {
	db *sqlx.DB
}

func NewShowsRepository(db *sqlx.DB) *ShowsRepository {
	return &ShowsRepository{db: db}
}

func (s *ShowsRepository) Add(ctx context.Context, show entities.Show) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO shows (show_id, dead_nation_id, number_of_tickets, start_time, title, venue)
				VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`,
		show.ShowID,
		show.DeadNationID,
		show.NumberOfTickets,
		show.StartTime,
		show.Title,
		show.Venue,
	)
	return err
}
