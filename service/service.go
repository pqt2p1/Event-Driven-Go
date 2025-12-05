package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"log/slog"
	stdHTTP "net/http"
	"tickets/db"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/message/event"
)

type Service struct {
	db              *sqlx.DB
	echoRouter      *echo.Echo
	watermillRouter *watermillMessage.Router
}

func New(
	dbConn *sqlx.DB,
	redisClient *redis.Client,
	spreadsheetsAPI event.SpreadsheetsAPI,
	receiptsService event.ReceiptsService,
	filesAPI event.FilesAPI,
) Service {
	ticketsRepo := db.NewTicketsRepository(dbConn)
	watermillLogger := watermill.NewSlogLogger(slog.Default())
	publisher, err := message.NewPublisher(redisClient, watermillLogger)
	if err != nil {
		panic(err)
	}
	eventBus := event.NewBus(publisher)
	eventsHandler := event.NewHandler(
		spreadsheetsAPI,
		receiptsService,
		ticketsRepo,
		filesAPI,
		eventBus,
	)
	eventProcessorConfig := event.NewProcessorConfig(redisClient, watermillLogger)
	echoRouter := ticketsHttp.NewHttpRouter(eventBus, ticketsRepo)
	watermillRouter := message.NewWatermillRouter(
		eventProcessorConfig,
		eventsHandler,
		watermillLogger,
	)

	return Service{
		db:              dbConn,
		echoRouter:      echoRouter,
		watermillRouter: watermillRouter,
	}
}

func (s Service) Run(ctx context.Context) error {
	if err := db.InitializeSchema(s.db); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

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
	return g.Wait()
}
