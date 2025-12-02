package message

import (
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
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
	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: redisClient,
	}, logger)
	if err != nil {
		return nil, err
	}

	// Wrap with CorrelationPublisherDecorator
	var publisher message.Publisher = pub
	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}

	return publisher, nil
}
