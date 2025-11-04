package workflows

import (
	"fmt"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/pkg/util"
	"go.temporal.io/sdk/workflow"
)

type CreateOrderWorkflowInput struct {
	OrderCode       string
	UserID          string
	Email           string
	Phone           string
	UserFullName    string
	EventID         string
	EventName       string
	Currency        string
	TotalAmount     int64
	Items           []CreateOrderItemInput
	PaymentProvider string
	RedirectUrl     string
	IdempotencyKey  string
}

type CreateOrderItemInput struct {
	OrderID         string
	TicketClassID   string
	TicketClassName string
	PriceAtPurchase int64
	Quantity        int32
	TotalAmount     int64
}

type CreateOrderWorkflowResult struct {
	PaymentUrl string
	Order      *models.Order
	OrderItems []models.OrderItem
}

func GetCreateOrderWorkflowID(oCode string) string {
	return fmt.Sprintf("CreateOrder:%s", oCode)
}

// CreateOrder handles order creation with saga pattern
// Workflow handles transactional saga:
// 1. Check Availability - Verify sufficient inventory
// 2. Create Order - Persist order record (saga begins)
// 3. Create Order Items - Persist order items (saga tracked)
// 4. Reserve Tickets - Lock inventory for 15 min (saga tracked)
// 5. Create Payment Intent - Generate payment URL
// Compensation on failure: Release inventory -> Delete items -> Delete order
func CreateOrder(ctx workflow.Context, in *CreateOrderWorkflowInput) (*CreateOrderWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting create order workflow", "orderCode", in.OrderCode)

	var compensations Compensations
	var err error

	defer func() {
		if err != nil {
			// activity failed, and workflow context is canceled
			logger.Error("Workflow failed, running compensations", "error", err)
			disconnectedCtx, _ := workflow.NewDisconnectedContext(ctx)
			compensations.Compensate(disconnectedCtx, false)
		}
	}()

	ctx = workflow.WithActivityOptions(ctx, getCreateOrderActivityOptions())

	// 1. Check availability
	available, err := checkAvailability(ctx, in)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, ErrInsufficientInventory
	}

	// 2. Create order
	o, err := createOrder(ctx, in)
	if err != nil {
		return nil, err
	}
	oID := o.ID.Hex()
	compensations.AddCompensation(oActs.DeleteOrder, oID)

	// 3. Create order items
	itms, err := createOrderItems(ctx, oID, in.Items)
	if err != nil {
		return nil, err
	}
	compensations.AddCompensation(oActs.DeleteOrderItems, oID)

	// 4. Reserve inventory
	expAt := util.TimeToISO8601Str(workflow.Now(ctx).Add(PaymentTimeout))
	err = reserveInventory(ctx, o.Code, expAt, in.Items)
	if err != nil {
		return nil, err
	}
	compensations.AddCompensation(iActs.ReleaseInventory, o.Code)

	// 5. Create payment intent
	pmtResp, err := processPayment(ctx, in)
	if err != nil {
		return nil, err
	}

	return &CreateOrderWorkflowResult{
		PaymentUrl: pmtResp.PaymentUrl,
		Order:      o,
		OrderItems: itms,
	}, nil
}
