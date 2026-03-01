package service

import (
	"context"

	"sample-app/dal"
	"sample-app/mq"
	"sample-app/rpc"
)

type OrderService struct {
	dal      *dal.OrderDAL
	rpcCli   *rpc.OrderClient
	producer *mq.Producer
}

func NewOrderService() *OrderService {
	return &OrderService{
		dal:      dal.NewOrderDAL(),
		rpcCli:   rpc.NewOrderClient(),
		producer: mq.NewProducer("order-events"),
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, productID int64, quantity int32, address string) (string, error) {
	// 1. Call downstream RPC to validate/create order
	orderID, err := s.rpcCli.CreateOrder(ctx, userID, productID, quantity, address)
	if err != nil {
		return "", err
	}

	// 2. Save to local database
	err = s.dal.Create(ctx, &dal.Order{
		ID:        orderID,
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Address:   address,
		Status:    "created",
	})
	if err != nil {
		return "", err
	}

	// 3. Send event to MQ
	_ = s.producer.SendMessage(ctx, orderID, []byte(`{"event":"order_created"}`))

	return orderID, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*dal.Order, error) {
	return s.dal.GetByID(ctx, orderID)
}
