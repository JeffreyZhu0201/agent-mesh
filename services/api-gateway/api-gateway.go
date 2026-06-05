package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Proxy to user-svc
	r.POST("/api/user/*path", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "user service proxy"})
	})

	// Proxy to agent-svc
	r.POST("/api/agent/*path", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "agent service proxy"})
	})

	// Proxy to plugin-svc
	r.POST("/api/plugin/*path", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "plugin service proxy"})
	})

	fmt.Println("Starting API Gateway at port 8080...")
	r.Run(":8080")
}
