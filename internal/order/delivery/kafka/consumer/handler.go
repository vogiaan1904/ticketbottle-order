package consumer

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/vogiaan1904/ticketbottle-order/internal/order/delivery/kafka"
	"github.com/vogiaan1904/ticketbottle-order/internal/order/service"
)

func (c *Consumer) HandlePaymentCompleted(ctx context.Context, msg *sarama.ConsumerMessage) error {
	c.l.Info(ctx, "HandlePaymentCompleted consumed")

	var e kafka.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		c.l.Error(ctx, "delivery.kafka.consumer.handlers.HandlePaymentCompleted: %v", err)
		return err
	}

	if err := c.svc.HandlePaymentCompleted(ctx, service.HandlePaymentCompletedInput{
		OrderCode: e.OrderCode,
	}); err != nil {
		c.l.Error(ctx, "delivery.kafka.consumer.handlers.HandlePaymentCompleted: %v", err)
		return err
	}

	return nil
}

func (c *Consumer) HandlePaymentFailed(ctx context.Context, msg *sarama.ConsumerMessage) error {
	c.l.Info(ctx, "HandlePaymentFailed consumed")

	var e kafka.PaymentFailedEvent
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		c.l.Error(ctx, "delivery.kafka.consumer.handlers.HandlePaymentFailed: %v", err)
		return err
	}

	if err := c.svc.HandlePaymentFailed(ctx, service.HandlePaymentFailedInput{
		OrderCode: e.OrderCode,
	}); err != nil {
		c.l.Error(ctx, "delivery.kafka.consumer.handlers.HandlePaymentFailed: %v", err)
		return err
	}

	return nil
}
