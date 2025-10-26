package main

import (
	"fmt"
	"net/http"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/handler"
	"github.com/gin-gonic/gin"
)

func ginHandlerAdapter(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h(c.Writer, c.Request)
	}
}

func main() {
	config.AppConfig = config.InitConfig()

	r := gin.Default()

	r.POST("/", ginHandlerAdapter(handler.CreateShortURL))
	r.GET("/:id", ginHandlerAdapter(handler.GetOriginalURL))

	fmt.Printf("Starting server on %s\n", config.AppConfig.ServerAddress)
	if err := r.Run(config.AppConfig.ServerAddress); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
