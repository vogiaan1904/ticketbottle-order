package grpc

import (
	"errors"

	orderpb "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/order"
)

var (
	ErrInvalidEventID       = errors.New("event_id is required")
	ErrInvalidUserID        = errors.New("user_id is required")
	ErrInvalidUserFullname  = errors.New("user_fullname is required")
	ErrInvalidUserEmail     = errors.New("user_email is required")
	ErrInvalidUserPhone     = errors.New("user_phone is required")
	ErrInvalidPaymentMethod = errors.New("payment_method is required")
	ErrInvalidCurrency      = errors.New("currency is required")
	ErrInvalidItems         = errors.New("items are required")
	ErrInvalidTicketClassID = errors.New("ticket_class_id is required")
	ErrInvalidQuantity      = errors.New("quantity must be greater than 0")
	ErrInvalidPage          = errors.New("page must be greater than 0")
	ErrInvalidPageSize      = errors.New("page_size must be greater than 0")
	ErrInvalidOrderID       = errors.New("order id is required")
	ErrInvalidFindOption    = errors.New("either code or id must be provided")
)

func (s *grpcService) validateCreateOrderRequest(req *orderpb.CreateOrderRequest) error {
	if req.GetEventId() == "" {
		return ErrInvalidEventID
	}
	if req.GetUserId() == "" {
		return ErrInvalidUserID
	}
	if req.GetUserFullname() == "" {
		return ErrInvalidUserFullname
	}
	if req.GetUserEmail() == "" {
		return ErrInvalidUserEmail
	}
	if req.GetUserPhone() == "" {
		return ErrInvalidUserPhone
	}
	if req.GetPaymentMethod() == "" {
		return ErrInvalidPaymentMethod
	}
	if req.GetCurrency() == "" {
		return ErrInvalidCurrency
	}
	if len(req.GetItems()) == 0 {
		return ErrInvalidItems
	}

	for _, item := range req.GetItems() {
		if err := validateCreateOrderItem(item); err != nil {
			return err
		}
	}

	return nil
}

func validateCreateOrderItem(item *orderpb.CreateOrderItem) error {
	if item.GetTicketClassId() == "" {
		return ErrInvalidTicketClassID
	}
	if item.GetQuantity() <= 0 {
		return ErrInvalidQuantity
	}

	return nil
}

func (s *grpcService) validateGetManyOrdersRequest(req *orderpb.GetManyOrdersRequest) error {
	if req.GetPage() <= 0 {
		return ErrInvalidPage
	}
	if req.GetPageSize() <= 0 {
		return ErrInvalidPageSize
	}

	return nil
}

func (s *grpcService) validateGetOrderRequest(req *orderpb.GetOrderRequest) error {
	if req.GetCode() == "" && req.GetId() == "" {
		return ErrInvalidFindOption
	}
	return nil
}

func (s *grpcService) validateListOrdersRequest(req *orderpb.ListOrdersRequest) error {
	// Filter is optional, no validation needed
	return nil
}

func (s *grpcService) validateCancelOrderRequest(req *orderpb.CancelOrderRequest) error {
	if req.GetId() == "" {
		return ErrInvalidOrderID
	}
	return nil
}
