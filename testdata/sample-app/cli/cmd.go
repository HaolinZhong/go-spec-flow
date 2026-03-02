package cli

import "sample-app/service"

// Simulates a cobra-style command pattern where function calls
// happen inside package-level variable function literals.

type Command struct {
	RunE func() error
}

var orderCmd = &Command{
	RunE: func() error {
		svc := service.NewOrderService()
		_, _ = svc.GetOrder(nil, "test")
		return nil
	},
}
