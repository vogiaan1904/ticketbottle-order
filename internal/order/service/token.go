package service

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
)

// func (s *implOrderService) checkoutTokenKey(ssID string) string {
// 	return fmt.Sprintf("checkout:token:%s", ssID)
// }

func (s *implOrderService) validateCheckoutToken(ctx context.Context, token string) (models.CheckoutTokenClaim, error) {
	p, err := s.jwt.Verify(token)
	if err != nil {
		return models.CheckoutTokenClaim{}, err
	}

	return p.CheckoutTokenClaim, nil
}
