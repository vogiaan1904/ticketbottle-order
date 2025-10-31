package activities

import (
	"context"
	"fmt"

	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/payment"
)

type PaymentActivities struct {
	Client payment.PaymentServiceClient
}

func NewPaymentActivities(client payment.PaymentServiceClient) *PaymentActivities {
	return &PaymentActivities{
		Client: client,
	}
}

func (a *PaymentActivities) CreatePaymentIntent(ctx context.Context, request *payment.CreatePaymentIntentRequest) (*payment.CreatePaymentIntentResponse, error) {
	resp, err := a.Client.CreatePaymentIntent(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return resp, nil
}

// CancelPayment cancels a payment (if supported by payment service)
// Note: This method may need to be implemented based on your payment service interface
func (a *PaymentActivities) CancelPayment(ctx context.Context, orderCode string, reason string) error {
	// TODO: Implement if payment service supports cancellation
	// For now, we'll just log the cancellation request
	return fmt.Errorf("payment cancellation not yet implemented for order: %s, reason: %s", orderCode, reason)
}
