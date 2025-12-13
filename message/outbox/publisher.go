package outbox

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"log/slog"
)

const outboxTopic = "events_to_forward"

func NewPublisherForDb(ctx context.Context, tx *sqlx.Tx) (message.Publisher, error) {
	logger := watermill.NewSlogLogger(slog.Default())

	sqlPublisher, err := watermillSQL.NewPublisher(
		tx,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	var publisher message.Publisher = sqlPublisher

	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}

	publisher = forwarder.NewPublisher(publisher, forwarder.PublisherConfig{
		ForwarderTopic: outboxTopic,
	})

	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}

	return publisher, nil
}

func PublishEventInTx(ctx context.Context, tx *sqlx.Tx, event interface{}) error {
	publisher, err := NewPublisherForDb(ctx, tx)
	if err != nil {
		return fmt.Errorf("could not create publisher: %w", err)
	}

	marshaler := cqrs.JSONMarshaler{
		GenerateName: cqrs.StructName,
	}

	eventBus, err := cqrs.NewEventBusWithConfig(
		publisher,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: marshaler,
		},
	)
	if err != nil {
		return fmt.Errorf("could not create event bus: %w", err)
	}

	return eventBus.Publish(ctx, event)
}
