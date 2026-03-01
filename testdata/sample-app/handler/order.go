package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"sample-app/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		svc: service.NewOrderService(),
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, c *app.RequestContext) {
	orderID, err := h.svc.CreateOrder(ctx, 1, 100, 2, "test address")
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(200, map[string]string{"order_id": orderID})
}

func (h *OrderHandler) GetOrder(ctx context.Context, c *app.RequestContext) {
	order, err := h.svc.GetOrder(ctx, "test-id")
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(200, order)
}
