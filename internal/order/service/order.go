package service

import (
	"context"
	"sync"
	"time"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/internal/order"
	repo "github.com/vogiaan1904/ticketbottle-order/internal/order/repository"
	ordWf "github.com/vogiaan1904/ticketbottle-order/internal/workflows/order"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/event"
	"github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
	"github.com/vogiaan1904/ticketbottle-order/pkg/mongo"
	"github.com/vogiaan1904/ticketbottle-order/pkg/util"
	"go.temporal.io/sdk/client"
)

func (s *implService) Create(ctx context.Context, in order.CreateOrderInput) (order.CreateOrderOutput, error) {
	var e *event.Event
	var eCfg *event.EventConfig

	wg := sync.WaitGroup{}
	var wgErr error

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
		return order.CreateOrderOutput{}, wgErr
	}

	if e.Status != event.EventStatus_EVENT_STATUS_PUBLISHED {
		s.l.Errorf(ctx, "internal.order.service.Create: %v", in.EventID)
		return order.CreateOrderOutput{}, order.ErrEventNotReadyForSale
	}

	if eCfg.AllowWaitRoom {
		claim, err := s.validateCheckoutToken(ctx, in.CheckoutToken)
		if err != nil {
			s.l.Errorf(ctx, "internal.order.service.Create: %v", err)
			return order.CreateOrderOutput{}, err
		}

		if claim.UserID != in.UserID || claim.EventID != in.EventID {
			s.l.Errorf(ctx, "internal.order.service.Create: %v", order.ErrInvalidCheckoutToken)
			return order.CreateOrderOutput{}, order.ErrInvalidCheckoutToken
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
		return order.CreateOrderOutput{}, err
	}

	if tcResp == nil || tcResp.GetTicketClasses() == nil || len(tcResp.GetTicketClasses()) == 0 {
		s.l.Errorf(ctx, "internal.order.service.Create: %v", in.EventID)
		return order.CreateOrderOutput{}, order.ErrTicketClassNotFound
	}

	for _, tc := range tcResp.GetTicketClasses() {
		tcMap[tc.GetId()] = tc
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
		return order.CreateOrderOutput{}, err
	}

	if !resp.Accept {
		s.l.Errorf(ctx, "internal.order.service.Create: %v", order.ErrNotEnoughTickets)
		return order.CreateOrderOutput{}, order.ErrNotEnoughTickets
	}

	// Calculate total amount and prepare order items
	amt := int64(0)
	itmIns := make([]repo.CreateOrderItemOption, len(in.Items))

	for i, itm := range in.Items {
		tc := tcMap[itm.TicketClassID]
		tt := tc.PriceCents * int64(itm.Quantity)
		amt += tt
		itmIns[i] = repo.CreateOrderItemOption{
			TicketClassID:   itm.TicketClassID,
			TicketClassName: tc.Name,
			PriceAtPurchase: tc.PriceCents,
			Quantity:        itm.Quantity,
			TotalAmount:     tt,
		}
	}

	// Generate order code and create order record
	code := util.GenerateOrderCodeWithEventPrefix(e.Name)

	o, err := s.repo.Create(ctx, repo.CreateOrderOption{
		Code:         code,
		UserID:       in.UserID,
		Email:        in.Email,
		Phone:        in.Phone,
		UserFullName: in.UserFullName,
		EventID:      in.EventID,
		Currency:     "VND",
		Status:       models.OrderStatusPending,
		TotalAmount:  amt,
	})
	if err != nil {
		s.l.Errorf(ctx, "failed to create order: %v", err)
		return order.CreateOrderOutput{}, err
	}

	itms, err := s.repo.CreateManyItems(ctx, o.ID.Hex(), itmIns)
	if err != nil {
		s.delete(ctx, o.ID.Hex())
		s.l.Errorf(ctx, "failed to create order items: %v", err)
		return order.CreateOrderOutput{}, err
	}

	// Start Temporal workflow for order processing
	wfOpts := client.StartWorkflowOptions{
		ID:        "create-order-" + o.Code,
		TaskQueue: s.cfg.TemporalTaskQueue,
	}

	// Prepare workflow parameters
	wfParams := ordWf.CreateOrderWorkflowParams{
		Order:           o,
		Items:           itms,
		PaymentProvider: string(in.PaymentMethod),
		RedirectUrl:     in.RedirectUrl,
		IdempotencyKey:  generatePaymentIdempotencyKey(o.Code, string(in.PaymentMethod)),
		TimeoutSeconds:  s.cfg.PaymentTimeoutSeconds,
	}

	wfRun, err := s.temporal.ExecuteWorkflow(ctx, wfOpts, ordWf.ProcessCreateOrderWorkflow, wfParams)
	if err != nil {
		s.l.Errorf(ctx, "failed to start pre-payment workflow: %v", err)
		s.delete(ctx, o.ID.Hex())
		return order.CreateOrderOutput{}, err
	}

	s.l.Infof(ctx, "Started pre-payment workflow for order %s, workflowID: %s, runID: %s", o.Code, wfRun.GetID(), wfRun.GetRunID())

	// Get the workflow result (payment URL)
	var wfRes ordWf.CreateOrderWorkflowResult
	err = wfRun.Get(ctx, &wfRes)
	if err != nil {
		s.l.Errorf(ctx, "pre-payment workflow failed: %v", err)
		return order.CreateOrderOutput{}, err
	}

	return order.CreateOrderOutput{
		Order:      o,
		OrderItems: itms,
		PaymentUrl: wfRes.PaymentUrl,
	}, nil
}

func (s *implService) confirm(ctx context.Context, code string) error {
	o, err := s.repo.GetOne(ctx, repo.GetOneOrderOption{
		FilterOrder: order.FilterOrder{
			Code: code,
		},
	})
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

	err = s.publishCheckoutCompletedEvent(ctx, order.PubCheckoutCompletedEventInput{
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
		s.l.Errorf(ctx, "internal.order.service.handlePaymentFailure.releaseTickets: %v", err)
	}

	o, err := s.repo.GetOne(ctx, repo.GetOneOrderOption{
		FilterOrder: order.FilterOrder{
			Code: code,
		},
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.l.Warnf(ctx, "internal.order.service.handlePaymentFailure.repo.GetByCode: %v", order.ErrOrderNotFound)
			return order.ErrOrderNotFound
		}
		s.l.Errorf(ctx, "internal.order.service.handlePaymentFailure.repo.GetByCode: %v", err)
		return err
	}

	_, err = s.repo.Update(ctx, o.ID.Hex(), repo.UpdateOrderOption{
		Status: models.OrderStatusPaymentFailed,
	})
	if err != nil {
		s.l.Errorf(ctx, "internal.order.service.handlePaymentFailure.repo.Update: %v", err)
		return err
	}

	err = s.publishCheckoutFailedEvent(ctx, order.PubCheckoutFailedEventInput{
		SessionID: o.SessionID,
		UserID:    o.UserID,
		EventID:   o.EventID,
	})
	if err != nil {
		s.l.Errorf(ctx, "internal.order.service.handlePaymentFailure.publishCheckoutFailedEvent: %v", err)
	}

	return nil
}

func (s *implService) delete(ctx context.Context, orderID string) error {
	err := s.repo.Delete(ctx, orderID)
	if err != nil {
		s.l.Errorf(ctx, "Failed to delete order %s: %v", orderID, err)
		return err
	}
	s.l.Infof(ctx, "Successfully deleted order %s", orderID)
	return nil
}

func (s *implService) Cancel(ctx context.Context, ID string) error {
	o, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.l.Warnf(ctx, "internal.order.service.Cancel: %v", order.ErrOrderNotFound)
			return order.ErrOrderNotFound
		}
		s.l.Errorf(ctx, "internal.order.service.Cancel.repo.GetByID:%v", err)
		return err
	}

	if o.Status != models.OrderStatusPending {
		s.l.Errorf(ctx, "internal.order.service.Cancel: %v", order.ErrOrderNotPending)
		return order.ErrOrderNotPending
	}

	if err := s.releaseTickets(ctx, o.Code); err != nil {
		s.l.Errorf(ctx, "internal.order.service.Cancel.releaseTickets: %v", err)
	}

	_, err = s.repo.Update(ctx, o.ID.Hex(), repo.UpdateOrderOption{
		Status: models.OrderStatusCancelled,
	})
	if err != nil {
		s.l.Errorf(ctx, "Failed to update order status to cancelled for %s: %v", o.Code, err)
		return order.ErrOrderCancellationFailed
	}

	if o.SessionID != "" {
		if err := s.publishCheckoutFailedEvent(ctx, order.PubCheckoutFailedEventInput{
			SessionID: o.SessionID,
			UserID:    o.UserID,
			EventID:   o.EventID,
		}); err != nil {
			s.l.Warnf(ctx, "Failed to publish checkout cancelled event for order %s: %v", o.Code, err)
		}
	}

	return nil
}

func (s *implService) GetMany(ctx context.Context, in order.GetManyOrderInput) (order.GetManyOrderOutput, error) {
	os, pag, err := s.repo.GetMany(ctx, repo.GetManyOrderOption(in))
	if err != nil {
		s.l.Errorf(ctx, "internal.order.service.GetMany.repo.GetMany: %v", err)
		return order.GetManyOrderOutput{}, err
	}

	return order.GetManyOrderOutput{
		Orders: os,
		Pag:    pag,
	}, nil
}

func (s *implService) GetByID(ctx context.Context, ID string) (models.Order, error) {
	o, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.l.Warnf(ctx, "internal.order.service.GetByID: %v", order.ErrOrderNotFound)
			return models.Order{}, order.ErrOrderNotFound
		}
		s.l.Errorf(ctx, "internal.order.service.GetByID.repo.GetByID:%v", err)
		return models.Order{}, err
	}

	return o, nil
}

func (s *implService) GetOne(ctx context.Context, in order.GetOneOrderInput) (models.Order, error) {
	o, err := s.repo.GetOne(ctx, repo.GetOneOrderOption(in))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.l.Warnf(ctx, "internal.order.service.GetOne: %v", order.ErrOrderNotFound)
			return models.Order{}, order.ErrOrderNotFound
		}
		s.l.Errorf(ctx, "internal.order.service.GetOne.repo.GetOne:%v", err)
		return models.Order{}, err
	}

	return o, nil
}

func (s *implService) List(ctx context.Context, in order.ListOrderInput) ([]models.Order, error) {
	os, err := s.repo.List(ctx, repo.ListOrderOption(in))
	if err != nil {
		s.l.Errorf(ctx, "internal.order.service.List.repo.List: %v", err)
		return nil, err
	}

	return os, nil
}
