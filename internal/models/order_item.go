package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ID              primitive.ObjectID `bson:"_id"`
	OrderID         primitive.ObjectID `bson:"order_id"`
	TicketClassID   string             `bson:"ticket_class_id"`
	TicketClassName string             `bson:"ticket_class_name"`
	PriceAtPurchase int64              `bson:"price_at_purchase"`
	Quantity        int32              `bson:"quantity"`
	TotalAmount     int64              `bson:"total_amount"`
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
	DeletedAt       *time.Time         `bson:"deleted_at,omitempty"`
}
