package service

import (
	"context"
	"fmt"

	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
)

func generatePaymentIdempotencyKey(orderCode string, provider string) string {
	return fmt.Sprintf("%s:%s", orderCode, provider)
}

func (s *implService) releaseTickets(ctx context.Context, code string) error {
	_, err := s.invSvc.Release(ctx, &inventory.ReleaseRequest{
		OrderCode: code,
	})
	if err != nil {
		s.l.Errorf(ctx, "Failed to release tickets for order %s: %v", code, err)
		return err
	}

	s.l.Infof(ctx, "Successfully released tickets for order %s", code)
	return nil
}
