package server

import "github.com/cloudwego/hertz/pkg/app"
import "github.com/cloudwego/hertz/pkg/route"

type Hertz struct {
	route.RouterGroup
}

func Default() *Hertz {
	return &Hertz{}
}

func New() *Hertz {
	return &Hertz{}
}

func (h *Hertz) Spin() {}

func (h *Hertz) Group(relativePath string, handlers ...app.HandlerFunc) *route.RouterGroup {
	return h.RouterGroup.Group(relativePath, handlers...)
}

func (h *Hertz) GET(relativePath string, handlers ...app.HandlerFunc) route.IRoutes {
	return h.RouterGroup.GET(relativePath, handlers...)
}

func (h *Hertz) POST(relativePath string, handlers ...app.HandlerFunc) route.IRoutes {
	return h.RouterGroup.POST(relativePath, handlers...)
}

func (h *Hertz) PUT(relativePath string, handlers ...app.HandlerFunc) route.IRoutes {
	return h.RouterGroup.PUT(relativePath, handlers...)
}

func (h *Hertz) DELETE(relativePath string, handlers ...app.HandlerFunc) route.IRoutes {
	return h.RouterGroup.DELETE(relativePath, handlers...)
}
