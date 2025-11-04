package workflows

import (
	"github.com/vogiaan1904/ticketbottle-order/internal/activities"
	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	repo "github.com/vogiaan1904/ticketbottle-order/internal/order/repository"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/payment"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func validateOrder(ctx workflow.Context, code string) (*models.Order, error) {
	var ord *models.Order
	err := workflow.ExecuteActivity(ctx, oActs.GetOrder, code).Get(ctx, &ord)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(
			"Failed to get order",
			"ORDER_NOT_FOUND",
			err,
		)
	}

	return ord, nil
}

func checkAvailability(ctx workflow.Context, in *CreateOrderWorkflowInput) (bool, error) {
	chkItms := make([]*inventory.CheckAvailabilityItem, len(in.Items))
	for i, itm := range in.Items {
		chkItms[i] = &inventory.CheckAvailabilityItem{
			TicketClassId: itm.TicketClassID,
			Quantity:      itm.Quantity,
		}
	}

	var available bool
	err := workflow.ExecuteActivity(ctx, iActs.CheckAvailability, chkItms).Get(ctx, &available)
	return available, err
}

func createOrder(ctx workflow.Context, in *CreateOrderWorkflowInput) (*models.Order, error) {
	opt := repo.CreateOrderOption{
		Code:         in.OrderCode,
		UserID:       in.UserID,
		Email:        in.Email,
		Phone:        in.Phone,
		UserFullName: in.UserFullName,
		EventID:      in.EventID,
		Currency:     in.Currency,
		Status:       models.OrderStatusPending,
		TotalAmount:  in.TotalAmount,
	}

	var o *models.Order
	err := workflow.ExecuteActivity(ctx, oActs.CreateOrder, opt).Get(ctx, &o)
	return o, err
}

func createOrderItems(ctx workflow.Context, oID string, ins []CreateOrderItemInput) ([]models.OrderItem, error) {
	opts := make([]repo.CreateOrderItemOption, len(ins))
	for i, itm := range ins {
		opts[i] = repo.CreateOrderItemOption{
			OrderID:         oID,
			TicketClassID:   itm.TicketClassID,
			TicketClassName: itm.TicketClassName,
		}
	}
	var itms []models.OrderItem

	err := workflow.ExecuteActivity(ctx, oActs.CreateOrderItems, oID, opts).Get(ctx, &itms)
	return itms, err
}

func reserveInventory(ctx workflow.Context, oCode string, expAt string, ins []CreateOrderItemInput) error {
	rsvItms := make([]*inventory.ReserveItem, len(ins))
	for i, itm := range ins {
		rsvItms[i] = &inventory.ReserveItem{
			TicketClassId: itm.TicketClassID,
			Quantity:      itm.Quantity,
		}
	}

	err := workflow.ExecuteActivity(ctx, iActs.ReserveInventory, oCode, expAt, rsvItms).Get(ctx, nil)
	return err
}

func updateOrderStatus(ctx workflow.Context, oID string, status models.OrderStatus) error {
	err := workflow.ExecuteActivity(ctx, oActs.UpdateOrderStatus, oID, status).Get(ctx, nil)
	return err
}

func processPayment(ctx workflow.Context, in *CreateOrderWorkflowInput) (*payment.CreatePaymentIntentResponse, error) {
	var resp *payment.CreatePaymentIntentResponse
	err := workflow.ExecuteActivity(ctx, pActs.CreatePaymentIntent,
		activities.CreatePaymentIntentInput{
			OrderCode:      in.OrderCode,
			TotalAmount:    in.TotalAmount,
			Currency:       in.Currency,
			Provider:       in.PaymentProvider,
			RedirectUrl:    in.RedirectUrl,
			IdempotencyKey: in.IdempotencyKey,
			TimeoutSeconds: int32(PaymentTimeout.Seconds()),
		},
	).Get(ctx, &resp)
	return resp, err
}

func confirmInventory(ctx workflow.Context, oCode string) error {
	err := workflow.ExecuteActivity(ctx, iActs.ConfirmInventory, oCode).Get(ctx, nil)
	return err
}
