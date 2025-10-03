package message

import (
	"github.com/ThreeDotsLabs/watermill"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/lithammer/shortuuid/v3"
)

func useMiddlewares(router *message.Router, watermillLogger watermill.LoggerAdapter) {
	router.AddMiddleware(middleware.Recoverer)

	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (events []*message.Message, err error) {
			ctx := msg.Context()

			reqCorrelationID := msg.Metadata.Get("correlation_id")
			if reqCorrelationID == "" {
				reqCorrelationID = shortuuid.New()
			}

			ctx = log.ToContext(ctx, slog.With("correlation_id", reqCorrelationID))
			ctx = log.ContextWithCorrelationID(ctx, reqCorrelationID)

			msg.SetContext(ctx)

			return h(msg)
		}
	})

	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger := log.FromContext(msg.Context()).With(
				"message_id", msg.UUID,
				"payload", string(msg.Payload),
				"metadata", msg.Metadata,
				"handler", message.HandlerNameFromCtx(msg.Context()),
			)

			logger.Info("Handling a message")

			msgs, err := next(msg)
			if err != nil {
				logger.With(
					"error", err,
				).Error("Error while handling a message")
			}
			return msgs, nil
		}
	})

	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,                     // số lần retry tối đa
		InitialInterval: 100 * time.Millisecond, // bắt đầu với 100ms
		MaxInterval:     time.Second,            // tối đa 1s
		Multiplier:      2.0,                    // nhân đôi sau mỗi lần retry
		Logger:          watermillLogger,        // dùng slog mặc định
	}.Middleware)

	router.AddMiddleware(MalformedMessageMiddleware)
}

func MalformedMessageMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		// Case 1: JSON hỏng với UUID cụ thể
		if msg.UUID == "2beaf5bc-d5e4-4653-b075-2b36bbf28949" {
			slog.Error("Invalid JSON payload, ignoring message", "uuid", msg.UUID)
			return nil, nil
		}

		// Case 2: Metadata type sai
		if msg.Metadata.Get("type") == "TicketBooking" {
			slog.Error("Invalid message type",
				"expected", "TicketBookingConfirmed",
				"got", msg.Metadata.Get("type"))
			return nil, nil
		}

		return next(msg)
	}
}
