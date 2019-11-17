package route

import (
	"log"
	"net/http"
)

type middleware func(http.Handler) http.Handler

type Router struct {
	middlewareChain [] middleware
	mux map[string]http.Handler
}

// 初始化router
func InitRouter() *Router{
	return &Router{
		middlewareChain: []middleware{},
		mux: map[string]http.Handler{},
	}
}

// 使用中间件
func (r *Router) Use(m middleware) {
	r.middlewareChain = append(r.middlewareChain, m)
}

// 添加路由
func (r *Router) Add(route string, h http.Handler) {
	var mergedHandler = h

	for i := len(r.middlewareChain) - 1; i >= 0; i-- {
		mergedHandler = r.middlewareChain[i](mergedHandler)
	}

	r.mux[route] = mergedHandler
}

// 添加func路由
func (r *Router) AddFuc(route string, f func(writer http.ResponseWriter, request *http.Request)) {
	r.Add(route, http.HandlerFunc(f))
}

// 启动服务
func (r *Router) ServeTLS(addr string, TSLCertFile string, TSLKeyFile string)  {
	for route, handler := range r.mux {
		http.Handle(route, handler)
	}
	
	log.Fatal("server listen: ", http.ListenAndServeTLS(
		addr,
		TSLCertFile,
		TSLKeyFile,
		nil))
}

// 启动http服务-test
func (r *Router) ServeHttp(addr string) {
	for route, handler := range r.mux {
		http.Handle(route, handler)
	}

	log.Fatal("server listen: ", http.ListenAndServe(addr, nil))
}