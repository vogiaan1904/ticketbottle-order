package service

type PaymentStatus string

const (
	PaymentStatusAuthorized PaymentStatus = "AUTHORIZED"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusExpired    PaymentStatus = "EXPIRED"
)

type HandlePaymentStatusInput struct {
	OrderCode string
	Status    PaymentStatus
}
