package order

import "github.com/vogiaan1904/ticketbottle-order/internal/models"

type CreateOrderWorkflowParams struct {
	Order           models.Order
	Items           []models.OrderItem
	PaymentProvider string
	RedirectUrl     string
	IdempotencyKey  string
	TimeoutSeconds  int32
}

type ConfirmOrderWorkflowParams struct {
	OrderCode string
	Status    models.OrderStatus
}

type CreateOrderWorkflowResult struct {
	PaymentUrl string
	OrderCode  string
}
