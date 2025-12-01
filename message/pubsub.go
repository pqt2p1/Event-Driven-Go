package message

import (
	eventLog "github.com/ThreeDotsLabs/go-event-driven/v2/common/log" // ← Thêm dòng nà
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"os"
)

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: os.Getenv(addr),
	})
}

func NewPublisher(redisClient *redis.Client, logger watermill.LoggerAdapter) (message.Publisher, error) {
	return redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: redisClient,
	}, logger)
}

func NewSubscriber(redisClient *redis.Client, group string, logger watermill.LoggerAdapter) (message.Subscriber, error) {
	return redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        redisClient,
		ConsumerGroup: group,
	}, logger)
}

func NewEventBus(redisClient *redis.Client, logger watermill.LoggerAdapter) (*cqrs.EventBus, error) {
	// 1. Tao publisher
	pub, err := NewPublisher(redisClient, logger)
	if err != nil {
		return nil, err
	}

	// 2. Wrap publisher voi correlation decorator
	var publisher message.Publisher = pub
	publisher = eventLog.CorrelationPublisherDecorator{publisher}

	// 3. Tao Marshaler
	marshaler := cqrs.JSONMarshaler{
		GenerateName: cqrs.StructName,
	}

	// 4. Tao EventBus
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
		return nil, err
	}

	return eventBus, nil
}
