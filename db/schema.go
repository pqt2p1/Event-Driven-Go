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
                );
                
                CREATE TABLE IF NOT EXISTS shows (
				  show_id UUID PRIMARY KEY,
				  dead_nation_id UUID NOT NULL,
				  number_of_tickets INT NOT NULL,
				  start_time TIMESTAMP NOT NULL,
				  title VARCHAR(255) NOT NULL,
				  venue VARCHAR(255) NOT NULL,
				
				  UNIQUE (dead_nation_id)
				);

				CREATE TABLE IF NOT EXISTS bookings (
					booking_id UUID PRIMARY KEY,
					show_id UUID NOT NULL,
					number_of_tickets INT NOT NULL,
					customer_email VARCHAR(255) NOT NULL,
					FOREIGN KEY (show_id) REFERENCES shows(show_id)
				);
        `
	_, err := db.Exec(schema)
	return err
}
