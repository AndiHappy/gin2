package gin

import "net/http"

// HandlerFunc defines the handler used by gin middleware as return value.
// HandlerFunc 定义了 gin 中间件使用的处理函数类型。
type HandlerFunc func(*Context)

// HandlersChain defines a HandlerFunc slice.
// 一个 HandlerFunc 切片类型，表示一组处理函数链。
type HandlersChain []HandlerFunc

type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	engine   *Engine
	root     bool
}

// Any implements IRoutes.
func (group *RouterGroup) Any(string, ...HandlerFunc) IRoutes {
	panic("unimplemented")
}

// DELETE implements IRoutes.
func (group *RouterGroup) DELETE(string, ...HandlerFunc) IRoutes {
	panic("unimplemented")
}

// GET implements IRoutes.
func (group *RouterGroup) GET(string, ...HandlerFunc) IRoutes {
	//TODO modify the path
	return group.returnObj()
}

// HEAD implements IRoutes.
func (group *RouterGroup) HEAD(string, ...HandlerFunc) IRoutes {
	panic("unimplemented")
}

// Handle implements IRoutes.
func (group *RouterGroup) Handle(string, string, ...HandlerFunc) IRoutes {
	panic("unimplemented")
}

// Match implements IRoutes.
func (group *RouterGroup) Match([]string, string, ...HandlerFunc) IRoutes {
	panic("unimplemented")
}

// OPTIONS implements IRoutes.
func (group *RouterGroup) OPTIONS(string, ...HandlerFunc) IRoutes {
	panic("unimplemented")
}

// PATCH implements IRoutes.
func (group *RouterGroup) PATCH(string, ...HandlerFunc) IRoutes {
	panic("unimplemented")
}

// POST implements IRoutes.
func (group *RouterGroup) POST(string, ...HandlerFunc) IRoutes {
	panic("unimplemented")
}

// PUT implements IRoutes.
func (group *RouterGroup) PUT(string, ...HandlerFunc) IRoutes {
	panic("unimplemented")
}

// Static implements IRoutes.
func (group *RouterGroup) Static(string, string) IRoutes {
	panic("unimplemented")
}

// StaticFS implements IRoutes.
func (group *RouterGroup) StaticFS(string, http.FileSystem) IRoutes {
	panic("unimplemented")
}

// StaticFile implements IRoutes.
func (group *RouterGroup) StaticFile(string, string) IRoutes {
	panic("unimplemented")
}

// StaticFileFS implements IRoutes.
func (group *RouterGroup) StaticFileFS(string, string, http.FileSystem) IRoutes {
	panic("unimplemented")
}

// Use adds middleware to the group, see example code in GitHub.
func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}

func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		//如果是root,返回的是Engine
		return group.engine
	}
	return group
}

// IRoutes defines all router handle interface.
type IRoutes interface {
	Use(...HandlerFunc) IRoutes

	Handle(string, string, ...HandlerFunc) IRoutes
	Any(string, ...HandlerFunc) IRoutes
	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
	DELETE(string, ...HandlerFunc) IRoutes
	PATCH(string, ...HandlerFunc) IRoutes
	PUT(string, ...HandlerFunc) IRoutes
	OPTIONS(string, ...HandlerFunc) IRoutes
	HEAD(string, ...HandlerFunc) IRoutes
	Match([]string, string, ...HandlerFunc) IRoutes

	StaticFile(string, string) IRoutes
	StaticFileFS(string, string, http.FileSystem) IRoutes
	Static(string, string) IRoutes
	StaticFS(string, http.FileSystem) IRoutes
}
