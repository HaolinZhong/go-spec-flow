package route

import "github.com/cloudwego/hertz/pkg/app"

type IRoutes interface {
	Group(relativePath string, handlers ...app.HandlerFunc) *RouterGroup
	GET(relativePath string, handlers ...app.HandlerFunc) IRoutes
	POST(relativePath string, handlers ...app.HandlerFunc) IRoutes
	PUT(relativePath string, handlers ...app.HandlerFunc) IRoutes
	DELETE(relativePath string, handlers ...app.HandlerFunc) IRoutes
}

type RouterGroup struct {
	basePath string
}

func (g *RouterGroup) Group(relativePath string, handlers ...app.HandlerFunc) *RouterGroup {
	return &RouterGroup{basePath: g.basePath + relativePath}
}

func (g *RouterGroup) GET(relativePath string, handlers ...app.HandlerFunc) IRoutes  { return g }
func (g *RouterGroup) POST(relativePath string, handlers ...app.HandlerFunc) IRoutes { return g }
func (g *RouterGroup) PUT(relativePath string, handlers ...app.HandlerFunc) IRoutes  { return g }
func (g *RouterGroup) DELETE(relativePath string, handlers ...app.HandlerFunc) IRoutes {
	return g
}
