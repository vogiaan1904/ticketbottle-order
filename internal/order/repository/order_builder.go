package repository

import (
	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *implRepository) buildOrderModel(opt CreateOrderOption) models.Order {
	now := r.clock()
	m := models.Order{
		ID:           r.db.NewObjectID(),
		SessionID:    opt.SessionID,
		Code:         opt.Code,
		UserID:       opt.UserID,
		UserFullName: opt.UserFullName,
		Email:        opt.Email,
		Phone:        opt.Phone,
		EventID:      opt.EventID,
		TotalAmount:  opt.TotalAmount,
		Status:       opt.Status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return m
}

func (r *implRepository) buildUpdateOrderModel(opt UpdateOrderOption) (models.Order, bson.M) {
	now := r.clock()

	m := opt.Model
	set := bson.M{
		"status": opt.Status,
	}
	m.Status = opt.Status

	if opt.PaidAt != nil {
		set["paid_at"] = *opt.PaidAt
		m.PaidAt = opt.PaidAt
	}

	m.UpdatedAt = now
	set["updated_at"] = now

	upDoc := bson.M{"$set": set}

	return m, upDoc
}
