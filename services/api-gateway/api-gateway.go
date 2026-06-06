package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func main() {
	r := gin.Default()

	// CORS middleware for browser access
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Tenant-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Service URLs from env vars (Docker service names) with localhost defaults
	userSvcAddr := getEnv("USER_SVC_URL", "http://user-svc:8081")
	agentSvcAddr := getEnv("AGENT_SVC_URL", "http://agent-svc:8082")
	pluginSvcAddr := getEnv("PLUGIN_SVC_URL", "http://plugin-svc:8083")

	userSvcURL, _ := url.Parse(userSvcAddr)
	agentSvcURL, _ := url.Parse(agentSvcAddr)
	pluginSvcURL, _ := url.Parse(pluginSvcAddr)

	r.Any("/api/user/*path", createProxy(userSvcURL, "/api/user"))
	r.Any("/api/agent/*path", createProxy(agentSvcURL, "/api/agent"))
	r.Any("/api/plugin/*path", createProxy(pluginSvcURL, "/api/plugin"))

	fmt.Printf("Starting API Gateway at port 8080...\n")
	fmt.Printf("  user-svc: %s\n", userSvcAddr)
	fmt.Printf("  agent-svc: %s\n", agentSvcAddr)
	fmt.Printf("  plugin-svc: %s\n", pluginSvcAddr)
	r.Run(":8080")
}

func createProxy(target *url.URL, basePath string) gin.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		// Get the wildcard path and prepend the base path
		path := c.Param("path")
		targetPath := basePath + path

		c.Request.URL.Path = targetPath
		c.Request.URL.RawPath = targetPath
		c.Request.Host = target.Host

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
