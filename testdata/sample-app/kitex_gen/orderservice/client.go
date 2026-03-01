package orderservice

import "context"

// Stub types mimicking Kitex generated code.
// gsf detects Kitex clients by checking if the receiver type
// originates from a kitex_gen package.

type CreateOrderRequest struct {
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	Address   string `json:"address"`
}

type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type GetOrderRequest struct {
	OrderID string `json:"order_id"`
}

type GetOrderResponse struct {
	OrderID   string `json:"order_id"`
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
	Status    string `json:"status"`
}

type Client interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error)
}

type client struct{}

func NewClient(serviceName string) Client {
	return &client{}
}

func (c *client) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	return &CreateOrderResponse{OrderID: "mock", Status: "created"}, nil
}

func (c *client) GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error) {
	return &GetOrderResponse{OrderID: req.OrderID, Status: "mock"}, nil
}
