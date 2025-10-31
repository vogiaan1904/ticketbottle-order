package order

import (
	"fmt"

	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
	"go.temporal.io/sdk/workflow"
)

// ProcessPrePaymentOrder handles the pre-payment phase of order processing
// This includes:
// 1. Validating the order exists
// 2. Reserving inventory
// 3. Creating payment intent
// 4. Setting up payment timeout monitoring
func ProcessCreateOrderWorkflow(ctx workflow.Context, params CreateOrderWorkflowParams) (*CreateOrderWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting create order workflow", "orderCode", params.Order.Code)

	rsvItms := make([]*inventory.ReserveItem, len(params.Items))
	for i, itm := range params.Items {
		rsvItms[i] = &inventory.ReserveItem{
			TicketClassId: itm.TicketClassID,
			Quantity:      itm.Quantity,
		}
	}

	expAt := workflow.Now(ctx).Add(PaymentTimeout).Format("2006-01-02T15:04:05Z07:00")

	if err := reserveInventory(ctx, params.Order.Code, expAt, rsvItms); err != nil {
		logger.Error("Failed to reserve inventory", "error", err)
		return nil, fmt.Errorf("failed to reserve inventory: %w", err)
	}
	logger.Info("Inventory reserved successfully", "orderCode", params.Order.Code)

	pmtResp, err := processPayment(ctx, params)
	if err != nil {
		logger.Error("Payment processing failed, releasing inventory", "error", err)
		if compensationErr := releaseInventory(ctx, params.Order.Code); compensationErr != nil {
			logger.Error("Compensation failed: could not release inventory", "error", compensationErr)
		}
		return nil, fmt.Errorf("payment failed: %w", err)
	}
	logger.Info("Payment intent created", "orderCode", params.Order.Code, "paymentUrl", pmtResp.PaymentUrl)

	return &CreateOrderWorkflowResult{
		PaymentUrl: pmtResp.PaymentUrl,
		OrderCode:  params.Order.Code,
	}, nil
}
