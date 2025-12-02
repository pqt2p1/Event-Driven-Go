package message

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"

	"tickets/message/event"
)

func NewWatermillRouter(receiptsService event.ReceiptsService, spreadsheetsAPI event.SpreadsheetsAPI, rdb *redis.Client, watermillLogger watermill.LoggerAdapter) *message.Router {
	router := message.NewDefaultRouter(watermillLogger)
	useMiddlewares(router, router.Logger())

	// Tao EventProcessor
	processorConfig := event.NewProcessorConfig(rdb, watermillLogger)
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, processorConfig)

	if err != nil {
		panic(err)
	}

	// Tao Handler
	handlers := event.NewHandler(spreadsheetsAPI, receiptsService)

	// Dang ky EventHandlers
	err = eventProcessor.AddHandlers(handlers.EventHandlers()...)
	if err != nil {
		panic(err)
	}

	return router
}
