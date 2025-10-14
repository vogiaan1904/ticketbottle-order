package kafka

type PaymentStatusEvent struct {
	OrderCode string `json:"order_code"`
	PaidAt    string `json:"paid_at,omitempty"`
	Status    string `json:"status"` // AUTHORIZED, FAILED, EXPIRED
}

type CheckoutCompletedEvent struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	EventID   string `json:"event_id"`
	Timestamp string `json:"timestamp"`
}

type CheckoutFailedEvent struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	EventID   string `json:"event_id"`
	Timestamp string `json:"timestamp"`
}
