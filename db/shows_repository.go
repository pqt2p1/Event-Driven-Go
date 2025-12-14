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

func (s *ShowsRepository) ShowByID(ctx context.Context, showID string) (entities.Show, error) {
	var show entities.Show
	err := s.db.GetContext(ctx, &show, `SELECT * FROM shows WHERE show_id = $1`, showID)
	if err != nil {
		return entities.Show{}, err
	}

	return show, nil
}
