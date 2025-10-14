package consumer

import (
	"context"
	"sync"

	"github.com/IBM/sarama"
	"github.com/vogiaan1904/ticketbottle-order/internal/order/delivery/kafka"
	"github.com/vogiaan1904/ticketbottle-order/internal/order/service"
	"github.com/vogiaan1904/ticketbottle-order/pkg/logger"
)

type Consumer struct {
	consGr sarama.ConsumerGroup
	svc    service.OrderService
	l      logger.Logger
	wg     sync.WaitGroup
}

func NewConsumer(
	consGr sarama.ConsumerGroup,
	svc service.OrderService,
	l logger.Logger,
) *Consumer {
	return &Consumer{
		consGr: consGr,
		svc:    svc,
		l:      l,
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	switch msg.Topic {
	case kafka.TopicPaymentStatus:
		return c.HandlePaymentStatus(ctx, msg)
	default:
		c.l.Warn(ctx, "Unknown topic", "topic", msg.Topic)
		return nil
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	topics := []string{kafka.TopicPaymentStatus}
	c.wg.Go(func() {
		for {
			if err := c.consGr.Consume(ctx, topics, c); err != nil {
				c.l.Errorf(ctx, "delivery.kafka.consumer.consumer.Start: %v", err)
			}

			if ctx.Err() != nil {
				c.l.Infof(ctx, "delivery.kafka.consumer.consumer.Start: %v", ctx.Err())
				return
			}
		}
	})

	// Handle errors
	c.wg.Go(func() {
		for err := range c.consGr.Errors() {
			c.l.Errorf(ctx, "delivery.kafka.consumer.consumer.Start: %v", err)
		}
	})

	c.l.Infof(ctx, "Consumer is consuming topics: %v", topics)
	return nil
}

func (c *Consumer) Close() error {
	if err := c.consGr.Close(); err != nil {
		return err
	}

	c.wg.Wait()
	return nil
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	c.l.Debug(context.Background(), "Consumer group session started")
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	c.l.Debug(context.Background(), "Consumer group session ended")
	return nil
}

func (c *Consumer) ConsumeClaim(ss sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			if err := c.processMessage(ss.Context(), message); err != nil {
				c.l.Error(ss.Context(), "delivery.kafka.consumer.consumer.ConsumeClaim: %v", err,
					"topic", message.Topic,
					"offset", message.Offset,
				)
				continue
			}

			ss.MarkMessage(message, "")

		case <-ss.Context().Done():
			return nil
		}
	}
}
