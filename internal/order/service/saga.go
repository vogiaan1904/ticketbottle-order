package service

import (
	"context"
)

func (s *implOrderService) compensate(ctx context.Context, saga *SagaCompensation) {
	s.l.Warnf(ctx, "Starting saga compensation/rollback")

	if saga.ItemsCreated && saga.CreatedOrder != nil {
		if err := s.itmSvc.DeleteByOrderID(ctx, saga.CreatedOrder.ID.Hex()); err != nil {
			s.l.Errorf(ctx, "Failed to delete order items during rollback: %v", err)
		}
	}

	if saga.CreatedOrder != nil {
		if err := s.deleteOrder(ctx, saga.CreatedOrder.ID.Hex()); err != nil {
			s.l.Errorf(ctx, "Failed to delete order during rollback: %v", err)
		}
	}

	if saga.TicketsReserved {
		s.releaseTickets(ctx, saga.CreatedOrder.Code)
	}
}
