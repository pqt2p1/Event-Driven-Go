package message

import (
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
