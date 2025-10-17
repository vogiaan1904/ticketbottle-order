package service

import "context"

type Service interface {
	Create(ctx context.Context, in CreateOrderInput) (CreateOrderOutput, error)
	Consumer
}

type Consumer interface {
	HandlePaymentCompleted(ctx context.Context, in HandlePaymentCompletedInput) error
	HandlePaymentFailed(ctx context.Context, in HandlePaymentFailedInput) error
}
