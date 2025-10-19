package service

import (
	"context"
	"sync"
	"time"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/internal/order"
	repo "github.com/vogiaan1904/ticketbottle-order/internal/order/repository"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/event"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/payment"
	"github.com/vogiaan1904/ticketbottle-order/pkg/util"
)

func (s *implService) Create(ctx context.Context, in CreateOrderInput) (CreateOrderOutput, error) {
	var e *event.Event
	var eCfg *event.EventConfig

	wg := sync.WaitGroup{}
	var wgErr error
	wg.Add(2)

	wg.Go(func() {
		resp, err := s.evSvc.FindOne(ctx, &event.FindOneEventRequest{
			Id: in.EventID,
		})
		if err != nil {
			s.l.Errorf(ctx, "internal.order.service.Create.evSvc.FindOne: %v", err)
			wgErr = err
			return
		}

		if resp.Event == nil {
			s.l.Errorf(ctx, "internal.order.service.Create: %v", in.EventID)
			wgErr = order.ErrEventNotFound
			return

		}

		e = resp.Event
	})

	wg.Go(func() {
		resp, err := s.evSvc.GetConfig(ctx, &event.GetEventConfigRequest{
			EventId: in.EventID,
		})
		if err != nil {
			s.l.Errorf(ctx, "internal.order.service.Create.evSvc.GetConfig: %v", err)
			wgErr = err
			return
		}

		if resp.EventConfig == nil {
			s.l.Errorf(ctx, "internal.order.service.Create: %v", in.EventID)
			wgErr = order.ErrEventConfigNotFound
			return
		}

		eCfg = resp.EventConfig
	})

	wg.Wait()
	if wgErr != nil {
		return CreateOrderOutput{}, wgErr
	}

	if e.Status != event.EventStatus_EVENT_STATUS_PUBLISHED {
		s.l.Errorf(ctx, "internal.order.service.Create: %v", in.EventID)
		return CreateOrderOutput{}, order.ErrEventNotReadyForSale
	}

	if eCfg.AllowWaitRoom {
		claim, err := s.validateCheckoutToken(ctx, in.CheckoutToken)
		if err != nil {
			s.l.Errorf(ctx, "internal.order.service.Create: %v", err)
			return CreateOrderOutput{}, err
		}

		if claim.UserID != in.UserID || claim.EventID != in.EventID {
			s.l.Errorf(ctx, "internal.order.service.Create: %v", order.ErrInvalidCheckoutToken)
			return CreateOrderOutput{}, order.ErrInvalidCheckoutToken
		}
	}

	tcMap := make(map[string]*inventory.TicketClass)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tcIds := make([]string, len(in.Items))
	for i, item := range in.Items {
		tcIds[i] = item.TicketClassID
	}

	tcResp, err := s.invSvc.FindManyTicketClass(ctx, &inventory.FindManyTicketClassRequest{
		EventId: in.EventID,
		Ids:     tcIds,
	})
	if err != nil {
		s.l.Errorf(ctx, "internal.order.service.Create.invSvc.FindManyTicketClass: %v", err)
		return CreateOrderOutput{}, err
	}

	if tcResp == nil || len(tcResp.TicketClasses) == 0 {
		s.l.Errorf(ctx, "internal.order.service.Create: %v", in.EventID)
		return CreateOrderOutput{}, order.ErrTicketClassNotFound
	}

	for _, tc := range tcResp.TicketClasses {
		tcMap[tc.Id] = tc
	}

	resp, err := s.invSvc.CheckAvailability(ctx, &inventory.CheckAvailabilityRequest{
		Items: func() []*inventory.CheckAvailabilityItem {
			itms := make([]*inventory.CheckAvailabilityItem, len(in.Items))
			for i, it := range in.Items {
				itms[i] = &inventory.CheckAvailabilityItem{
					TicketClassId: it.TicketClassID,
					Quantity:      it.Quantity,
				}
			}
			return itms
		}(),
	})
	if err != nil {
		s.l.Errorf(ctx, "internal.order.service.Create.invSvc.CheckAvailability: %v", err)
		return CreateOrderOutput{}, err
	}

	if !resp.Accept {
		s.l.Errorf(ctx, "internal.order.service.Create: %v", order.ErrNotEnoughTickets)
		return CreateOrderOutput{}, order.ErrNotEnoughTickets
	}

	saga := &SagaCompensation{}
	defer func() {
		if r := recover(); r != nil {
			s.compensate(ctx, saga)
			panic(r)
		}
	}()

	code := util.GenerateOrderCodeWithEventPrefix(e.Name)

	_, err = s.invSvc.Reserve(ctx, &inventory.ReserveRequest{
		OrderCode: code,
		Items: func() []*inventory.ReserveItem {
			itms := make([]*inventory.ReserveItem, len(in.Items))
			for i, it := range in.Items {
				itms[i] = &inventory.ReserveItem{
					TicketClassId: it.TicketClassID,
					Quantity:      it.Quantity,
				}
			}
			return itms
		}(),
	})
	if err != nil {
		s.l.Errorf(ctx, "failed to reserve tickets: %v", err)
		return CreateOrderOutput{}, err
	}
	saga.TicketsReserved = true

	amt := int64(0)
	itmIns := make([]repo.CreateOrderItemOption, len(in.Items))

	for _, i := range in.Items {
		tc := tcMap[i.TicketClassID]
		tt := tc.PriceCents * int64(i.Quantity)
		amt += tt
		itmIns = append(itmIns, repo.CreateOrderItemOption{
			TicketClassID:   i.TicketClassID,
			TicketClassName: tc.Name,
			PriceAtPurchase: tc.PriceCents,
			Quantity:        i.Quantity,
			TotalAmount:     tt,
		})
	}

	o, err := s.repo.Create(ctx, repo.CreateOrderOption{
		Code:         code,
		UserID:       in.UserID,
		Email:        in.Email,
		UserFullName: in.UserFullName,
		EventID:      in.EventID,
		Currency:     "VND", // VND only for now
		TotalAmount:  amt,
	})
	if err != nil {
		s.releaseTickets(ctx, code)
		s.l.Errorf(ctx, "failed to create order: %v", err)
		return CreateOrderOutput{}, err
	}
	saga.CreatedOrder = &o

	itms, err := s.repo.CreateManyItems(ctx, o.ID.Hex(), itmIns)
	if err != nil {
		s.releaseTickets(ctx, code)
		s.deleteOrder(ctx, o.ID.Hex())
		s.l.Errorf(ctx, "failed to create order items: %v", err)
		return CreateOrderOutput{}, err
	}
	saga.ItemsCreated = true

	pResp, err := s.pmtSvc.CreatePaymentIntent(ctx, &payment.CreatePaymentIntentRequest{
		OrderCode:      o.Code,
		AmountCents:    o.TotalAmount,
		Currency:       "VND",
		Provider:       payment.PaymentProvider(payment.PaymentProvider_value[string(in.PaymentMethod)]),
		RedirectUrl:    in.RedirectUrl,
		IdempotencyKey: generatePaymentIdempotencyKey(o.Code, string(in.PaymentMethod)),
		TimeoutSeconds: s.cfg.PaymentTimeoutSeconds,
	})
	if err != nil {
		s.compensate(ctx, saga)
		s.l.Errorf(ctx, "failed to create payment intent: %v", err)
		return CreateOrderOutput{}, err
	}

	return CreateOrderOutput{
		Order:       o,
		OrderItems:  itms,
		RedirectUrl: pResp.PaymentUrl,
	}, nil
}

func (s *implService) confirm(ctx context.Context, code string) error {
	o, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		s.l.Error(ctx, "internal.order.service.HandlePaymentStatus: cannot get order by code %v: %v", code, err)
		return err
	}

	if o.Status == models.OrderStatusCompleted {
		s.l.Warnf(ctx, "Order %s is already confirmed", o.Code)
		return nil // Idempotent - already processed
	}

	_, err = s.invSvc.Confirm(ctx, &inventory.ConfirmRequest{
		OrderCode: code,
	})
	if err != nil {
		s.l.Errorf(ctx, "Failed to confirm reservation for order %s: %v", code, err)
		return err
	}

	_, err = s.repo.Update(ctx, o.ID.Hex(), repo.UpdateOrderOption{
		Status: models.OrderStatusCompleted,
	})
	if err != nil {
		s.l.Errorf(ctx, "Failed to update order status for %s: %v", code, err)
		return err
	}

	err = s.publishCheckoutCompletedEvent(ctx, PubCheckoutCompletedEventInput{
		SessionID: o.SessionID,
		UserID:    o.UserID,
		EventID:   o.EventID,
	})
	if err != nil {
		s.l.Errorf(ctx, "Failed to publish order confirmed event for %s: %v", code, err)
	}

	return nil
}

func (s *implService) handlePaymentFailure(ctx context.Context, code string) error {
	err := s.releaseTickets(ctx, code)
	if err != nil {
		s.l.Errorf(ctx, "Failed to release tickets for failed payment order %s: %v", code, err)
	}

	o, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		s.l.Error(ctx, "internal.order.service.HandlePaymentStatus: cannot get order by code %v: %v", code, err)
		return err
	}

	_, err = s.repo.Update(ctx, o.ID.Hex(), repo.UpdateOrderOption{
		Status: models.OrderStatusPaymentFailed,
	})
	if err != nil {
		s.l.Errorf(ctx, "Failed to update order status to cancelled for %s: %v", code, err)
		return err
	}

	err = s.publishCheckoutFailedEvent(ctx, PubCheckoutFailedEventInput{
		SessionID: o.SessionID,
		UserID:    o.UserID,
		EventID:   o.EventID,
	})
	if err != nil {
		s.l.Errorf(ctx, "Failed to publish order cancelled event for %s: %v", code, err)
	}

	return nil
}

func (s *implService) deleteOrder(ctx context.Context, orderID string) error {
	err := s.repo.Delete(ctx, orderID)
	if err != nil {
		s.l.Errorf(ctx, "Failed to delete order %s: %v", orderID, err)
		return err
	}
	s.l.Infof(ctx, "Successfully deleted order %s", orderID)
	return nil
}
