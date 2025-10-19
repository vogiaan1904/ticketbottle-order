package service

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/order"
)

func (s *implService) HandlePaymentCompleted(ctx context.Context, in order.HandlePaymentCompletedInput) error {
	if err := s.confirm(ctx, in.OrderCode); err != nil {
		s.l.Errorf(ctx, "service.consumer.HandlePaymentCompleted: %v", err)
		return err
	}

	return nil
}

func (s *implService) HandlePaymentFailed(ctx context.Context, in order.HandlePaymentFailedInput) error {
	if err := s.handlePaymentFailure(ctx, in.OrderCode); err != nil {
		s.l.Errorf(ctx, "service.consumer.HandlePaymentFailed: %v", err)
		return err
	}

	return nil
}
