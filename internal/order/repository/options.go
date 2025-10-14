package repository

import (
	"github.com/vogiaan1904/ticketbottle-order/internal/models"
)

type CreateOrderOption struct {
	Session      string
	Code         string
	UserID       string
	UserFullName string
	Email        string
	EventID      string
	TotalAmount  int64
	Currency     string
	Status       models.OrderStatus
}

type UpdateOrderOption struct {
	Status models.OrderStatus
}
