package order

import (
	"fmt"

	"github.com/vogiaan1904/ticketbottle-order/internal/activities"
	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/payment"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func validateOrder(ctx workflow.Context, code string) (*models.Order, error) {
	var ord *models.Order
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getOrderActivityOptions()),
		(*activities.OrderActivities).GetOrder,
		code,
	).Get(ctx, &ord)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(
			"Failed to get order",
			"ORDER_NOT_FOUND",
			err,
		)
	}

	return ord, nil
}

func reserveInventory(ctx workflow.Context, orderCode string, expiresAt string, items []*inventory.ReserveItem) error {
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getInventoryActivityOptions()),
		(*activities.InventoryActivities).ReserveInventory,
		orderCode,
		expiresAt,
		items,
	).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to reserve inventory: %w", err)
	}
	return nil
}

func updateOrderStatus(ctx workflow.Context, orderID string, status models.OrderStatus) error {
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getOrderActivityOptions()),
		(*activities.OrderActivities).UpdateOrderStatus,
		orderID,
		status,
	).Get(ctx, nil)
	return err
}

func processPayment(ctx workflow.Context, params CreateOrderWorkflowParams) (*payment.CreatePaymentIntentResponse, error) {
	req := &payment.CreatePaymentIntentRequest{
		OrderCode:      params.Order.Code,
		AmountCents:    params.Order.TotalAmount,
		Currency:       params.Order.Currency,
		Provider:       payment.PaymentProvider(payment.PaymentProvider_value[params.PaymentProvider]),
		RedirectUrl:    params.RedirectUrl,
		IdempotencyKey: params.IdempotencyKey,
		TimeoutSeconds: params.TimeoutSeconds,
	}

	var resp *payment.CreatePaymentIntentResponse
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getPaymentActivityOptions()),
		(*activities.PaymentActivities).CreatePaymentIntent,
		req,
	).Get(ctx, &resp)
	return resp, err
}

func confirmInventory(ctx workflow.Context, orderCode string) error {
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getInventoryActivityOptions()),
		(*activities.InventoryActivities).ConfirmInventory,
		orderCode,
	).Get(ctx, nil)
	return err
}
