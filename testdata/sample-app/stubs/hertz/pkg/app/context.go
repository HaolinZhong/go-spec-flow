package app

import "context"

type HandlerFunc func(ctx context.Context, c *RequestContext)

type RequestContext struct{}

func (c *RequestContext) JSON(code int, obj any) {}
func (c *RequestContext) String(code int, s string) {}
