package main

import (
	"fmt"
	"net/http"

	"github.com/Rusich90/shurl.git/internal/handler"
	"github.com/gin-gonic/gin"
)

func ginHandlerAdapter(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h(c.Writer, c.Request)
	}
}

func main() {
	r := gin.Default()

	r.POST("/", ginHandlerAdapter(handler.CreateShortURL))
	r.GET("/:id", ginHandlerAdapter(handler.GetOriginalURL))

	fmt.Println("Starting server on :8080")
	if err := r.Run(); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
