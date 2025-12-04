package main

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"tickets/message"

	"tickets/adapters"
	"tickets/service"
)

func main() {
	log.Init(slog.LevelInfo)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	apiClients, err := clients.NewClients(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	redisClient := message.NewRedisClient("REDIS_ADDR")
	defer redisClient.Close()

	database, err := sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}
	defer database.Close()

	spreadsheetsAPI := adapters.NewSpreadsheetsAPIClient(apiClients)
	receiptsService := adapters.NewReceiptsServiceClient(apiClients)
	filesAPI := adapters.NewFilesAPIClient(apiClients)

	err = service.New(
		database,
		redisClient,
		spreadsheetsAPI,
		receiptsService,
		filesAPI,
	).Run(ctx)
	if err != nil {
		panic(err)
	}
}
