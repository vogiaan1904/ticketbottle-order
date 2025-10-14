package consumer

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/vogiaan1904/ticketbottle-order/internal/order/delivery/kafka"
	"github.com/vogiaan1904/ticketbottle-order/internal/order/service"
)

func (c *Consumer) HandlePaymentStatus(ctx context.Context, msg *sarama.ConsumerMessage) error {
	c.l.Info(ctx, "HandlePaymentStatus consumed")

	var e kafka.PaymentStatusEvent
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		c.l.Error(ctx, "delivery.kafka.consumer.handlers.HandleCheckoutCompleted: %v", err)
		return err
	}

	if err := c.svc.HandlePaymentStatus(ctx, service.HandlePaymentStatusInput{
		OrderCode: e.OrderCode,
		Status:    service.PaymentStatus(e.Status),
	}); err != nil {
		c.l.Error(ctx, "delivery.kafka.consumer.handlers.HandleCheckoutCompleted: %v", err)
		return err
	}

	return nil
}
