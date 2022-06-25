package main

import (
	"net/http"

	"github.com/pedrogao/log"
	"github.com/pedrogao/web"
)

func main() {
	engine := web.New()

	engine.GET("/", func(ctx *web.Context) {
		ctx.String(http.StatusOK, "hello world")
	})

	if err := engine.Run(":3000"); err != nil {
		log.Fatalf("start server err: %s", err)
	}
}
