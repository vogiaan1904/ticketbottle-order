package repository

import (
	"time"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
)

type CreateOrderOption struct {
	SessionID    string
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
	Model  models.Order
	Status models.OrderStatus
	PaidAt *time.Time
}
