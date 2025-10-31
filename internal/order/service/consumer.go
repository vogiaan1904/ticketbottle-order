package service

import (
	"context"

	"github.com/vogiaan1904/ticketbottle-order/internal/models"
	"github.com/vogiaan1904/ticketbottle-order/internal/order"
	ordWf "github.com/vogiaan1904/ticketbottle-order/internal/workflows/order"
	"go.temporal.io/sdk/client"
)

func (s *implService) HandlePaymentCompleted(ctx context.Context, in order.HandlePaymentCompletedInput) error {
	s.l.Infof(ctx, "Handling payment completed for order: %s", in.OrderCode)

	wfOpts := client.StartWorkflowOptions{
		ID:        "order-postpayment-" + in.OrderCode,
		TaskQueue: s.cfg.TemporalTaskQueue,
	}

	wfParams := ordWf.ConfirmOrderWorkflowParams{
		OrderCode: in.OrderCode,
		Status:    models.OrderStatusCompleted,
	}

	wfRun, err := s.temporal.ExecuteWorkflow(ctx, wfOpts, ordWf.ProcessConfirmOrderWorkflow, wfParams)
	if err != nil {
		s.l.Errorf(ctx, "Failed to start post-payment workflow: %v", err)
		return err
	}

	s.l.Infof(ctx, "Started post-payment workflow for order %s, workflowID: %s, runID: %s", in.OrderCode, wfRun.GetID(), wfRun.GetRunID())

	err = wfRun.Get(ctx, nil)
	if err != nil {
		s.l.Errorf(ctx, "Post-payment workflow failed: %v", err)
		return err
	}

	s.l.Infof(ctx, "Payment completed successfully for order: %s", in.OrderCode)
	return nil
}

func (s *implService) HandlePaymentFailed(ctx context.Context, in order.HandlePaymentFailedInput) error {
	s.l.Infof(ctx, "Handling payment failed for order: %s", in.OrderCode)

	err := s.handlePaymentFailure(ctx, in.OrderCode)
	if err != nil {
		s.l.Errorf(ctx, "Failed to handle payment failure: %v", err)
		return err
	}

	s.l.Infof(ctx, "Payment failure handled successfully for order: %s", in.OrderCode)
	return nil
}
