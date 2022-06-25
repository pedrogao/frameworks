package web

import (
	"net/http"

	"github.com/pedrogao/log"
)

type router struct {
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{handlers: map[string]HandlerFunc{}}
}

func (r *router) addRoute(method, pattern string, handler HandlerFunc) {
	log.Infof("Route %4s - %s", method, pattern)
	key := method + "-" + pattern
	r.handlers[key] = handler
}

func (r *router) handle(c *Context) {
	key := c.Method + "-" + c.Path
	defer func() {
		log.Infof("handle route, path: %s, method: %s, status: %d",
			c.Path, c.Method, c.StatusCode)
	}()

	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		c.Stringf(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
