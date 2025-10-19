package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID            primitive.ObjectID `bson:"_id"`
	SessionID     string             `bson:"session_id,omitempty"`
	Code          string             `bson:"code"`
	UserID        string             `bson:"user_id"`
	UserFullName  string             `bson:"user_full_name"`
	Email         string             `bson:"email"`
	EventID       string             `bson:"event_id"`
	TotalAmount   int64              `bson:"total_amount"`
	Currency      string             `bson:"currency"`
	PaymentMethod PaymentMethod      `bson:"payment_method"`
	Status        OrderStatus        `bson:"status"`
	PaidAt        *time.Time         `bson:"paid_at,omitempty"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
	DeletedAt     *time.Time         `bson:"deleted_at,omitempty"`
}

type OrderStatus string

const (
	OrderStatusPending       OrderStatus = "PENDING"
	OrderStatusTimeout       OrderStatus = "TIMEOUT"
	OrderStatusCompleted     OrderStatus = "COMPLETED"
	OrderStatusCancelled     OrderStatus = "CANCELLED"
	OrderStatusPaymentFailed OrderStatus = "PAYMENT_FAILED"
	OrderStatusRefunded      OrderStatus = "REFUNDED"
)

type PaymentMethod string

const (
	PaymentMethodVNPAY   PaymentMethod = "VNPAY"
	PaymentMethodZalopay PaymentMethod = "ZALOPAY"
	PaymentMethodPayOS   PaymentMethod = "PAYOS"
)
