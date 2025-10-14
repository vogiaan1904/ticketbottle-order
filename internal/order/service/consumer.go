package service

import (
	"context"
)

func (s *implOrderService) HandlePaymentStatus(ctx context.Context, in HandlePaymentStatusInput) error {
	switch in.Status {
	case PaymentStatusAuthorized:
		return s.confirm(ctx, in.OrderCode)
	case PaymentStatusFailed, PaymentStatusExpired:
		return s.handlePaymentFailure(ctx, in.OrderCode)
	}

	s.l.Warnf(ctx, "Unknown payment status: %s for order: %s", in.Status, in.OrderCode)

	return nil
}
