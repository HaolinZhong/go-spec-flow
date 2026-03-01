package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	"sample-app/handler"
)

func Register(h *server.Hertz) {
	orderHandler := handler.NewOrderHandler()

	api := h.Group("/api")
	v1 := api.Group("/v1")

	v1.POST("/orders", orderHandler.CreateOrder)
	v1.GET("/orders/:id", orderHandler.GetOrder)
}
