package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"os"
	"sync"
	"testing"
	"tickets/entities"
)

var db *sqlx.DB
var getDbOnce sync.Once

func getDb() *sqlx.DB {
	getDbOnce.Do(func() {
		var err error
		db, err = sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
		if err != nil {
			panic(err)
		}
	})
	return db
}
func TestAdd_Idempotent(t *testing.T) {
	db := getDb()
	repo := NewTicketsRepository(db)
	ctx := context.Background()

	ticket := entities.Ticket{
		TicketID: uuid.NewString(),
		Price:    entities.Money{Amount: "100", Currency: "USD"},
	}

	err := repo.Add(ctx, ticket)
	if err != nil {
		t.Fatal()
	}

	err = repo.Add(ctx, ticket)
	if err != nil {
		t.Fatal()
	}

	tickets, _ := repo.FindAll(ctx)
	if len(tickets) != 1 {
		t.Fatal()
	}

}

func TestMain(m *testing.M) {
	db := getDb()
	err := InitializeSchema(db)
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
