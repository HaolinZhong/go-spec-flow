package rpc

import (
	"context"

	"sample-app/kitex_gen/orderservice"
)

// OrderClient wraps the Kitex generated client.
type OrderClient struct {
	cli orderservice.Client
}

func NewOrderClient() *OrderClient {
	return &OrderClient{
		cli: orderservice.NewClient("order-service"),
	}
}

func (c *OrderClient) CreateOrder(ctx context.Context, userID, productID int64, quantity int32, address string) (string, error) {
	resp, err := c.cli.CreateOrder(ctx, &orderservice.CreateOrderRequest{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Address:   address,
	})
	if err != nil {
		return "", err
	}
	return resp.OrderID, nil
}

func (c *OrderClient) GetOrder(ctx context.Context, orderID string) (*orderservice.GetOrderResponse, error) {
	return c.cli.GetOrder(ctx, &orderservice.GetOrderRequest{OrderID: orderID})
}
