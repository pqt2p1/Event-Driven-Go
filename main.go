package main

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
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

	spreadsheetsAPI := adapters.NewSpreadsheetsAPIClient(apiClients)
	receiptsService := adapters.NewReceiptsServiceClient(apiClients)

	err = service.New(
		redisClient,
		spreadsheetsAPI,
		receiptsService,
	).Run(ctx)
	if err != nil {
		panic(err)
	}
}
