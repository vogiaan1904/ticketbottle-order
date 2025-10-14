package service

import "github.com/vogiaan1904/ticketbottle-order/internal/models"

type CreateOrderInput struct {
	CheckoutToken   string
	UserID          string
	Email           string
	UserFullName    string
	EventID         string
	Currency        string
	RedirectUrl     string
	PaymentProvider models.PaymentProvider
	Items           []OrderItemInput
}

type CreateOrderOutput struct {
	OrderCode   string
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
