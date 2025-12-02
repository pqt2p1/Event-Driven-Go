package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitializeSchema(db *sqlx.DB) error {
	schema := `
                CREATE TABLE IF NOT EXISTS tickets (
                        ticket_id UUID PRIMARY KEY,
                        price_amount NUMERIC(10, 2) NOT NULL,
                        price_currency CHAR(3) NOT NULL,
                        customer_email VARCHAR(255) NOT NULL
                )
        `
	_, err := db.Exec(schema)
	return err
}
