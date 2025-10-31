package order

import (
	"fmt"
	"time"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"go.temporal.io/sdk/workflow"
)

// ProcessPostPaymentOrder handles the post-payment phase of order processing
// This is triggered after successful payment and includes:
// 1. Validating order status
// 2. Updating order status to payment success
// 3. Confirming inventory reservation
// 4. Finalizing the order
func ProcessConfirmOrderWorkflow(ctx workflow.Context, params ConfirmOrderWorkflowParams) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting confirm order workflow", "orderCode", params.OrderCode)

	o, err := validateOrder(ctx, params.OrderCode)
	if err != nil {
		return fmt.Errorf("failed to validate order: %w", err)
	}

	if o.Status != models.OrderStatusPending {
		logger.Warn("Order already processed", "orderCode", params.OrderCode, "status", o.Status)
		if o.Status == models.OrderStatusCompleted {
			// Idempotent - already processed successfully
			return nil
		}
		return ErrOrderAlreadyProcessed
	}

	logger.Info("Payment successful, confirming inventory", "orderCode", params.OrderCode)

	if err := confirmInventory(ctx, params.OrderCode); err != nil {
		logger.Error("Failed to confirm inventory", "error", err)
		// This is a critical error - we need manual intervention
		// TODO: Implement manual intervention task creation or alerting
		return fmt.Errorf("failed to confirm inventory: %w", err)
	}
	logger.Info("Inventory confirmed successfully", "orderCode", params.OrderCode)

	if err := updateOrderStatus(ctx, o.ID.Hex(), models.OrderStatusCompleted); err != nil {
		// Retry once after a short delay
		logger.Warn("Failed to update order status, retrying", "error", err)
		_ = workflow.Sleep(ctx, time.Second*5)
		if retryErr := updateOrderStatus(ctx, o.ID.Hex(), models.OrderStatusCompleted); retryErr != nil {
			return fmt.Errorf("failed to update order status after retry: %w", retryErr)
		}
	}
	logger.Info("Order completed successfully", "orderCode", params.OrderCode)

	return nil
}
