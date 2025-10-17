package service

type HandlePaymentCompletedInput struct {
	OrderCode string
}

type HandlePaymentFailedInput struct {
	OrderCode string
}
