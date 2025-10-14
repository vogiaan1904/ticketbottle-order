package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ID              primitive.ObjectID  `bson:"_id"`
	OrderID         primitive.ObjectID  `bson:"order_id"`
	TicketClassID   string              `bson:"ticket_class_id"`
	TicketClassName string              `bson:"ticket_class_name"`
	PriceAtPurchase int64               `bson:"price_at_purchase"`
	Quantity        int32               `bson:"quantity"`
	CreatedAt       primitive.DateTime  `bson:"created_at"`
	UpdatedAt       primitive.DateTime  `bson:"updated_at"`
	DeletedAt       *primitive.DateTime `bson:"deleted_at,omitempty"`
}
