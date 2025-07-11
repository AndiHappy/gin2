package gin

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sync"
)

// abortIndex represents a typical value used in abort functions.
const abortIndex int8 = math.MaxInt8 >> 1

// Context is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
// Context 是 gin 中最重要的部分。它允许我们在中间件之间传递变量，管理流程，验证请求的 JSON，
// 并渲染 JSON 响应等。是一个自定义的组件内容
type Context struct {
	writermem responseWriter
	Request   *http.Request
	Writer    ResponseWriter

	// Key-Value pairs for the context
	Params   Params
	handlers HandlersChain
	index    int8
	fullPath string

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[any]any

	// Errors is a list of errors attached to all the handlers/middlewares who used this context.
	Errors errorMsgs

	// Accepted defines a list of manually accepted formats for content negotiation.
	Accepted []string

	// queryCache caches the query result from c.Request.URL.Query().
	queryCache url.Values

	// formCache caches c.Request.PostForm, which contains the parsed form data from POST, PATCH,
	// or PUT body parameters.
	formCache url.Values

	// SameSite allows a server to define a cookie attribute making it impossible for
	// the browser to send this cookie along with cross-site requests.
	sameSite http.SameSite

	// params 不对外暴露的情况
	params *Params

	// Context has Engine as a field to access the engine's methods and properties.
	engine *Engine

	// skippedNodes is a slice of skippedNode, which contains information about nodes that were skipped during routing.
	// skippedNodes 是一个 skippedNode 的切片，包含了在路由过程中跳过的节点的信息。
	// 这个字段用于存储在处理请求时跳过的节点信息。
	skippedNodes *[]skippedNode
}

// 设置json的数据
func (c *Context) JSON(code int, obj any) {
	// c.Render(code, render.JSON{Data: obj})
	panic("unimplemented")
}

/************************************/
/********** CONTEXT CREATION ********/
/************************************/

func (c *Context) reset() {
	c.Writer = &c.writermem
	c.Params = c.Params[:0]
	c.handlers = nil
	c.index = -1

	c.fullPath = ""
	c.Keys = nil
	c.Errors = c.Errors[:0]
	c.Accepted = nil
	c.queryCache = nil
	c.formCache = nil
	c.sameSite = 0
	*c.params = (*c.params)[:0]
	*c.skippedNodes = (*c.skippedNodes)[:0]
}

// Copy returns a copy of the current context that can be safely used outside the request's scope.
// This has to be used when the context has to be passed to a goroutine.
func (c *Context) Copy() *Context {
	cp := Context{
		writermem: c.writermem,
		Request:   c.Request,
		engine:    c.engine,
	}

	cp.writermem.ResponseWriter = nil
	cp.Writer = &cp.writermem
	cp.index = abortIndex
	cp.handlers = nil
	cp.fullPath = c.fullPath

	cKeys := c.Keys
	cp.Keys = make(map[any]any, len(cKeys))
	c.mu.RLock()
	for k, v := range cKeys {
		cp.Keys[k] = v
	}
	c.mu.RUnlock()

	cParams := c.Params
	cp.Params = make([]Param, len(cParams))
	copy(cp.Params, cParams)

	return &cp
}

// HandlerName returns the main handler's name. For example if the handler is "handleGetUsers()",
// this function will return "main.handleGetUsers".
func (c *Context) HandlerName() string {
	panic("unimplemented")
}

// HandlerNames returns a list of all registered handlers for this context in descending order,
// following the semantics of HandlerName()
func (c *Context) HandlerNames() []string {
	panic("unimplemented")
	// hn := make([]string, 0, len(c.handlers))
	// for _, val := range c.handlers {
	// 	if val == nil {
	// 		continue
	// 	}
	// 	hn = append(hn, nameOfFunction(val))
	// }
	// return hn
}

// Handler returns the main handler.
func (c *Context) Handler() HandlerFunc {
	panic("unimplemented")
	// return c.handlers.Last()
}

// FullPath returns a matched route full path. For not found routes
// returns an empty string.
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    c.FullPath() == "/user/:id" // true
//	})
func (c *Context) FullPath() string {
	return c.fullPath
}

/************************************/
/*********** FLOW CONTROL ***********/
/************************************/

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in GitHub.
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		if c.handlers[c.index] != nil {
			c.handlers[c.index](c)
		}
		c.index++
	}
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// Abort prevents pending handlers from being called. Note that this will not stop the current handler.
// Let's say you have an authorization middleware that validates that the current request is authorized.
// If the authorization fails (ex: the password does not match), call Abort to ensure the remaining handlers
// for this request are not called.
func (c *Context) Abort() {
	c.index = abortIndex
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// For example, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (c *Context) AbortWithStatus(code int) {
	panic("unimplemented")
	// c.Status(code)
	// c.Writer.WriteHeaderNow()
	// c.Abort()
}

// AbortWithStatusJSON calls `Abort()` and then `JSON` internally.
// This method stops the chain, writes the status code and return a JSON body.
// It also sets the Content-Type as "application/json".
func (c *Context) AbortWithStatusJSON(code int, jsonObj any) {
	panic("unimplemented")
	// c.Abort()
	// c.JSON(code, jsonObj)
}

// AbortWithError calls `AbortWithStatus()` and `Error()` internally.
// This method stops the chain, writes the status code and pushes the specified error to `c.Errors`.
// See Context.Error() for more details.
func (c *Context) AbortWithError(code int, err error) *Error {
	c.AbortWithStatus(code)
	return c.Error(err)
}

/************************************/
/********* ERROR MANAGEMENT *********/
/************************************/

// Error attaches an error to the current context. The error is pushed to a list of errors.
// It's a good idea to call Error for each error that occurred during the resolution of a request.
// A middleware can be used to collect all the errors and push them to a database together,
// print a log, or append it in the HTTP response.
// Error will panic if err is nil.
func (c *Context) Error(err error) *Error {
	if err == nil {
		panic("err is nil")
	}

	var parsedError *Error
	ok := errors.As(err, &parsedError)
	if !ok {
		parsedError = &Error{
			Err:  err,
			Type: ErrorTypePrivate,
		}
	}

	c.Errors = append(c.Errors, parsedError)
	return parsedError
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key any, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[any]any)
	}

	c.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (c *Context) Get(key any) (value any, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Keys[key]
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (c *Context) MustGet(key any) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic(fmt.Sprintf("key %v does not exist", key))
}
