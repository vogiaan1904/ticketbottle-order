package order

import (
	"github.com/vogiaan1904/ticketbottle-order/internal/activities"
	"go.temporal.io/sdk/workflow"
)

// releaseInventory releases reserved inventory (compensation action)
func releaseInventory(ctx workflow.Context, orderCode string) error {
	return workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getInventoryActivityOptions()),
		(*activities.InventoryActivities).ReleaseInventory,
		orderCode,
	).Get(ctx, nil)
}

// cancelPayment cancels a payment (compensation action)
func cancelPayment(ctx workflow.Context, orderCode string, reason string) error {
	return workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getPaymentActivityOptions()),
		(*activities.PaymentActivities).CancelPayment,
		orderCode,
		reason,
	).Get(ctx, nil)
}

// deleteOrder deletes an order (compensation action)
func deleteOrder(ctx workflow.Context, orderID string) error {
	return workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getOrderActivityOptions()),
		(*activities.OrderActivities).DeleteOrder,
		orderID,
	).Get(ctx, nil)
}

// deleteOrderItems deletes order items (compensation action)
func deleteOrderItems(ctx workflow.Context, orderID string) error {
	return workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getOrderActivityOptions()),
		(*activities.OrderActivities).DeleteOrderItems,
		orderID,
	).Get(ctx, nil)
}
