package gee

import (
	"log"
	"net/http"
	"strings"
)

// HandlerFunc defines the request handler used by gee   HandlerFunc定义了gee使用的请求处理程序
type HandleFunc func(c *Context)

// Engine implement the interface of ServeHTTP  引擎实现了ServeHTTP的接口
type (
	RouterGroup struct {
		prefix      string       //前缀
		middlewares []HandleFunc // support middleware 中间件
		parent      *RouterGroup // support nesting    支持嵌套
		engine      *Engine      // all groups share a Engine instance  所有组共享一个Engine实例
	}

	Engine struct {
		*RouterGroup
		router *router
		groups []*RouterGroup // store all groups 存储所有组
	}
)

// New is the constructor of gee.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group is defined to create a new RouterGroup  Group  定义创建一个新的RouterGroup
// remember all groups share the same Engine instance  所有组共享同一个Engine实例
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandleFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandleFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandleFunc) {
	group.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// Use is defined to add middleware to the group
func (group *RouterGroup) Use(middlewares ...HandleFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//确认是否存在中间件
	var middlewares []HandleFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	//构造请求——将请求封装到Context中
	c := newContext(w, req)
	//将中间件封装到Context中
	c.handlers = middlewares
	//执行请求
	engine.router.handle(c)
}
