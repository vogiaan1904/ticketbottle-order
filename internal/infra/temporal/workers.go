package temporal

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const (
	CreateOrderTaskQueue  = "create-order-task-queue"
	ConfirmOrderTaskQueue = "confirm-order-task-queue"
)

func NewCreateOrderWorker(tprCli client.Client) worker.Worker {
	return worker.New(tprCli, CreateOrderTaskQueue, worker.Options{
		MaxConcurrentWorkflowTaskExecutionSize: 100,
		MaxConcurrentActivityExecutionSize:     100,
	})
}

func NewConfirmOrderWorker(tprCli client.Client) worker.Worker {
	return worker.New(tprCli, ConfirmOrderTaskQueue, worker.Options{
		MaxConcurrentWorkflowTaskExecutionSize: 50,
		MaxConcurrentActivityExecutionSize:     50,
	})
}
