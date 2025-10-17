package service

import "github.com/vogiaan1904/ticketbottle-order/internal/models"

type CreateOrderInput struct {
	CheckoutToken string
	UserID        string
	Email         string
	UserFullName  string
	EventID       string
	Currency      string
	RedirectUrl   string
	PaymentMethod models.PaymentMethod
	Items         []OrderItemInput
}

type CreateOrderOutput struct {
	Order       models.Order
	OrderItems  []models.OrderItem
	RedirectUrl string
}

type OrderItemInput struct {
	TicketClassID string
	Quantity      int32
}

type SagaCompensation struct {
	CreatedOrder    *models.Order
	ItemsCreated    bool
	TicketsReserved bool
}

type ReservedTicket struct {
	OrderCode     string
	TicketClassID string
	Quantity      int32
	ReservationID string
}
