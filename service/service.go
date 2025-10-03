package service

import (
	"context"
	"errors"
	"github.com/ThreeDotsLabs/watermill"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"log/slog"
	stdHTTP "net/http"
	"tickets/adapters"
	ticketsHttp "tickets/http"
	"tickets/message"
)

type Service struct {
	echoRouter      *echo.Echo
	watermillRouter *watermillMessage.Router
}

func New(
	redisClient *redis.Client,
	spreadsheetsAPI *adapters.SpreadsheetsAPIClient,
	receiptsService *adapters.ReceiptsServiceClient,
) Service {
	watermillLogger := watermill.NewSlogLogger(slog.Default())
	publisher, _ := message.NewPublisher(redisClient, watermillLogger)
	echoRouter := ticketsHttp.NewHttpRouter(publisher, spreadsheetsAPI)
	watermillRouter := message.NewWatermillRouter(
		receiptsService,
		spreadsheetsAPI,
		redisClient,
		watermillLogger,
	)

	return Service{
		echoRouter:      echoRouter,
		watermillRouter: watermillRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	// Goroutine 1: watermill router
	g.Go(func() error {
		// dùng ctx thay vì context.Background()
		return s.watermillRouter.Run(ctx)
	})

	// Goroutine 2: echo server
	g.Go(func() error {
		<-s.watermillRouter.Running()

		err := s.echoRouter.Start(":8080")
		if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
			return err
		}
		return nil
	})

	// Goroutine 3: shutdown echo khi ctx bị hủy
	g.Go(func() error {
		<-ctx.Done()
		return s.echoRouter.Shutdown(ctx)
	})

	// Chờ tất cả goroutine xong
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
