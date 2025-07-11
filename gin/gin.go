package gin

import (
	"html/template"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const ginSupportMinGoVer = 21

var defaultPlatform string

const defaultMultipartMemory = 32 << 20 // 32 MB
const escapedColon = "\\:"
const colon = ":"
const backslash = "\\"

type Engine struct {
	RouterGroup

	// 下面是引擎的配置选项
	// RedirectTrailingSlash enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// RedirectFixedPath if enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// HandleMethodNotAllowed if enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// ForwardedByClientIP if enabled, client IP will be parsed from the request's headers that
	// match those stored at `(*gin.Engine).RemoteIPHeaders`. If no IP was
	// fetched, it falls back to the IP obtained from
	// `(*gin.Context).Request.RemoteAddr`.
	ForwardedByClientIP bool

	// AppEngine was deprecated.
	// Deprecated: USE `TrustedPlatform` WITH VALUE `gin.PlatformGoogleAppEngine` INSTEAD
	// #726 #755 If enabled, it will trust some headers starting with
	// 'X-AppEngine...' for better integration with that PaaS.
	AppEngine bool

	// UseRawPath if enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool

	// UnescapePathValues if true, the path value will be unescaped.
	// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
	// as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool

	// RemoveExtraSlash a parameter can be parsed from the URL even with extra slashes.
	// See the PR #1817 and issue #1644
	RemoveExtraSlash bool

	// RemoteIPHeaders list of headers used to obtain the client IP when
	// `(*gin.Engine).ForwardedByClientIP` is `true` and
	// `(*gin.Context).Request.RemoteAddr` is matched by at least one of the
	// network origins of list defined by `(*gin.Engine).SetTrustedProxies()`.
	RemoteIPHeaders []string

	// TrustedPlatform if set to a constant of value gin.Platform*, trusts the headers set by
	// that platform, for example to determine the client IP
	TrustedPlatform string

	// MaxMultipartMemory value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64

	// UseH2C enable h2c support.
	// UseH2C indicates whether to enable HTTP/2 Cleartext (h2c) support for the server.
	// When set to true, the server will accept HTTP/2 connections over cleartext (non-TLS) TCP connections.
	// This is useful for environments where TLS termination is handled by a reverse proxy or load balancer.
	UseH2C bool

	// ContextWithFallback enable fallback Context.Deadline(), Context.Done(), Context.Err() and Context.Value() when Context.Request.Context() is not nil.
	ContextWithFallback bool

	// FuncMap is a map of functions that can be used in templates.
	FuncMap template.FuncMap

	allNoRoute  HandlersChain
	allNoMethod HandlersChain
	noRoute     HandlersChain
	noMethod    HandlersChain

	pool sync.Pool

	// trees is a slice of methodTree, each methodTree contains a method and its corresponding route tree.
	// trees 是一个 methodTree 的切片，每个 methodTree 包含一个方法及其对应的路由树。
	// 每个 methodTree 的 root 节点是一个 node 类型，表示路由树的根节点。
	trees methodTrees

	maxParams uint16

	maxSections uint16

	//代理相关的内容
	trustedProxies []string
	trustedCIDRs   []*net.IPNet
}

// ServeHTTP implements http.Handler.
// 这个是拦截net。http包的ServeHTTP方法，是 Gin 的核心方法之一，负责处理 HTTP 请求。
// 它实现了 http.Handler 接口，接收 http.ResponseWriter 和 *http.Request 作为参数。
// 很多基于net/http 包的框架都是实现这个方法来拦截 HTTP 请求。
// 大多数Go语言的Web框架（如Echo、Beego、Fiber等）确实都采用了类似Gin的方式处理HTTP请求，即底层基于net/http包，并通过实现ServeHTTP方法作为请求的入口点。这是Go标准库提供的标准HTTP处理模式，框架在此基础上进行扩展和封装，提供更高级的功能和更便捷的API。
// 这种设计模式的主要优势包括：
// 与标准库完全兼容
// 可以复用net/http的成熟功能（如连接管理、请求解析等）
// 方便与其他基于net/http的中间件和工具集成
// 不同框架可能在路由实现、中间件机制等方面有所差异，但核心的请求处理流程都是通过ServeHTTP方法实现的。
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//sync.Pool是Go语言中用于管理临时对象池的工具,这个是一个高性能的对象池实现，用于减少内存分配和垃圾回收的开销。
	//现在对象池中存储的是*Context类型的对象，这个对象是Gin框架中用于处理HTTP请求的上下文对象。
	c := engine.pool.Get().(*Context)
	//初始化Context对象
	c.writermem.reset(w)
	c.Request = req
	c.reset()

	engine.handleHTTPRequest(c)

	// 处理完请求后，将Context对象放回对象池中，以便下次复用
	engine.pool.Put(c)
}

// handleHTTPRequest is the main HTTP request handler for the Gin engine.
// 处理HTTP请求的主要方法，负责根据请求方法和路径找到相应的路由，并执行对应的处理函数。
func (engine *Engine) handleHTTPRequest(c *Context) {
	httpMethod := c.Request.Method
	rPath := c.Request.URL.Path

	unescape := false
	if engine.UseRawPath && len(c.Request.URL.RawPath) > 0 {
		//RawPath是URL结构体中的一个可选字段，用于存储路径的原始转义形式
		//例如: 原始路径为/foo%2fbar，解码后为/foo/bar，此时RawPath会被设置为/foo%2fbar
		rPath = c.Request.URL.RawPath
		unescape = engine.UnescapePathValues
	}

	// If the path is empty, set it to "/"
	if engine.RemoveExtraSlash {
		rPath = cleanPath(rPath)
	}

	// Find root of the tree for the given HTTP method
	t := engine.trees
	// Traverse all route trees and find the tree that matches the request method
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}
		root := t[i].root
		// Find route in tree
		value := root.getValue(rPath, c.params, c.skippedNodes, unescape)
		if value.params != nil {
			c.Params = *value.params
		}
		if value.handlers != nil {
			c.handlers = value.handlers
			c.fullPath = value.fullPath
			c.Next()
			c.writermem.WriteHeaderNow()
			return
		}
		if httpMethod != http.MethodConnect && rPath != "/" {
			// if value.tsr && engine.RedirectTrailingSlash {
			// 	redirectTrailingSlash(c)
			// 	return
			// }
			// if engine.RedirectFixedPath && redirectFixedPath(c, root, engine.RedirectFixedPath) {
			// 	return
			// }
		}
		break
	}

	// if engine.HandleMethodNotAllowed && len(t) > 0 {
	// 	// According to RFC 7231 section 6.5.5, MUST generate an Allow header field in response
	// 	// containing a list of the target resource's currently supported methods.
	// 	allowed := make([]string, 0, len(t)-1)
	// 	for _, tree := range engine.trees {
	// 		if tree.method == httpMethod {
	// 			continue
	// 		}
	// 		if value := tree.root.getValue(rPath, nil, c.skippedNodes, unescape); value.handlers != nil {
	// 			allowed = append(allowed, tree.method)
	// 		}
	// 	}
	// 	if len(allowed) > 0 {
	// 		c.handlers = engine.allNoMethod
	// 		c.writermem.Header().Set("Allow", strings.Join(allowed, ", "))
	// 		serveError(c, http.StatusMethodNotAllowed, default405Body)
	// 		return
	// 	}
	// }
	//
	// c.handlers = engine.allNoRoute
	// serveError(c, http.StatusNotFound, default404Body)
}

// OptionFunc defines the function to change the default configuration
type OptionFunc func(*Engine)

// Default returns an Engine instance with the Logger and Recovery middleware already attached.
// 返回一个带有 Logger 和 Recovery 中间件的 Engine 实例。
func Default(opts ...OptionFunc) *Engine {
	debugPrintWARNINGDefault()
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine.With(opts...)
}

// Use attaches a global middleware to the router. i.e. the middleware attached through Use() will be
// included in the handlers chain for every single request. Even 404, 405, static files...
// For example, this is the right place for a logger or error management middleware.
// 使用 Use() 方法可以将全局中间件附加到路由器上。
func (engine *Engine) Use(middleware ...HandlerFunc) IRoutes {
	engine.RouterGroup.Use(middleware...)
	engine.rebuild404Handlers()
	engine.rebuild405Handlers()
	return engine
}

func (engine *Engine) rebuild405Handlers() {
	engine.allNoRoute = engine.combineHandlers(engine.noRoute)
}

func (engine *Engine) rebuild404Handlers() {
	engine.allNoMethod = engine.combineHandlers(engine.noMethod)
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	assert1(finalSize < int(abortIndex), "too many handlers")
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

// isTrustedProxy will check whether the IP address is included in the trusted list according to Engine.trustedCIDRs
func (engine *Engine) isTrustedProxy(ip net.IP) bool {
	if engine.trustedCIDRs == nil {
		return false
	}
	for _, cidr := range engine.trustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// isUnsafeTrustedProxies checks if Engine.trustedCIDRs contains all IPs, it's not safe if it has (returns true)
func (engine *Engine) isUnsafeTrustedProxies() bool {
	return engine.isTrustedProxy(net.ParseIP("0.0.0.0")) || engine.isTrustedProxy(net.ParseIP("::"))
}

// updateRouteTree do update to the route tree recursively
func updateRouteTree(n *node) {
	n.path = strings.ReplaceAll(n.path, escapedColon, colon)
	n.fullPath = strings.ReplaceAll(n.fullPath, escapedColon, colon)
	n.indices = strings.ReplaceAll(n.indices, backslash, colon)
	if n.children == nil {
		return
	}
	for _, child := range n.children {
		updateRouteTree(child)
	}
}

// updateRouteTrees do update to the route trees
func (engine *Engine) updateRouteTrees() {
	for _, tree := range engine.trees {
		updateRouteTree(tree.root)
	}
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) Run(addr ...string) (err error) {
	defer func() { debugPrintError(err) }()

	if engine.isUnsafeTrustedProxies() {
		debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
			"Please check https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	}
	engine.updateRouteTrees()
	address := resolveAddress(addr)
	debugPrint("Listening and serving HTTP on %s\n", address)
	// 这里要求的是Engine的Handler方法,也就是Handler接口
	err = http.ListenAndServe(address, engine.Handler())
	return
}

// Handler returns the Engine instance as an http.Handler.
// go的组合模式，实际上是将Engine作为http.Handler接口的实现
func (engine *Engine) Handler() http.Handler {
	if !engine.UseH2C { //默认的情况下，engine.UseH2C=false,则返回 engine
		return engine
	}

	// 如果engine.UseH2C为true，则返回一个h2c的Handler
	// h2c是HTTP/2 Cleartext的缩写，表示在不使用TLS的情况下支持HTTP/2协议。
	// 这通常用于在内部网络或测试环境中使用HTTP/2，而不需要加密。
	// 通过h2c.NewHandler(engine, h2s)创建一个新的h2c处理器，将engine作为处理器传递给它。
	// 这样，当接收到HTTP/2请求时，h2c处理器将调用engine来处理请求。
	// 这使得engine能够处理HTTP/2请求，同时仍然保持与HTTP/1.x的兼容性。
	h2s := &http2.Server{}
	return h2c.NewHandler(engine, h2s)
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}

// New returns a new blank Engine instance without any middleware attached.
// By default, the configuration is:
// - RedirectTrailingSlash:  true
// - RedirectFixedPath:      false
// - HandleMethodNotAllowed: false
// - ForwardedByClientIP:    true
// - UseRawPath:             false
// - UnescapePathValues:     true
func New(opts ...OptionFunc) *Engine {
	debugPrintWARNINGNew()
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		// FuncMap:                template.FuncMap{},
		// RedirectTrailingSlash:  true,
		// RedirectFixedPath:      false,
		// HandleMethodNotAllowed: false,
		// ForwardedByClientIP:    true,
		// RemoteIPHeaders:        []string{"X-Forwarded-For", "X-Real-IP"},
		// TrustedPlatform:        defaultPlatform,
		// UseRawPath:             false,
		// RemoveExtraSlash:       false,
		// UnescapePathValues:     true,
		// MaxMultipartMemory:     defaultMultipartMemory,
		// trees:                  make(methodTrees, 0, 9),
		// delims:                 render.Delims{Left: "{{", Right: "}}"},
		// secureJSONPrefix:       "while(1);",
		// trustedProxies:         []string{"0.0.0.0/0", "::/0"},
		// trustedCIDRs:           defaultTrustedCIDRs,
	}
	engine.engine = engine
	engine.pool.New = func() any {
		return engine.allocateContext(engine.maxParams)
	}
	return engine.With(opts...)
}

// With returns a Engine with the configuration set in the OptionFunc.
func (engine *Engine) With(opts ...OptionFunc) *Engine {
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

func (engine *Engine) allocateContext(maxParams uint16) *Context {
	v := make(Params, 0, maxParams)
	skippedNodes := make([]skippedNode, 0, engine.maxSections)
	return &Context{engine: engine, params: &v, skippedNodes: &skippedNodes}
}

func getMinVer(v string) (uint64, error) {
	first := strings.IndexByte(v, '.')
	last := strings.LastIndexByte(v, '.')
	if first == last {
		return strconv.ParseUint(v[first+1:], 10, 64)
	}
	return strconv.ParseUint(v[first+1:last], 10, 64)
}

func debugPrintWARNINGNew() {
	debugPrint(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

`)
}

func debugPrintWARNINGDefault() {
	if v, e := getMinVer(runtime.Version()); e == nil && v < ginSupportMinGoVer {
		debugPrint(`[WARNING] Now Gin requires Go 1.23+.`)
	}
	debugPrint(`[WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.`)
}
