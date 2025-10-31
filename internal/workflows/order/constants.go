package order

import "time"

const (
	// PaymentTimeout is the duration to wait for payment completion
	// 5 minutes for payment + 1 minute buffer for callback processing
	PaymentTimeout = 6 * time.Minute

	// SignalNamePaymentCompleted is the signal name for payment completion
	SignalNamePaymentCompleted = "payment-completed"

	// SignalNamePaymentFailed is the signal name for payment failure
	SignalNamePaymentFailed = "payment-failed"

	// TaskQueueName is the task queue name for order workflows
	TaskQueueName = "order-workflows"
)
