package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var configFile = flag.String("f", "user-svc.yaml", "the config file")

type Config struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

func main() {
	flag.Parse()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"users": []string{}})
	})

	fmt.Printf("Starting User Service at port %d...\n", 8081)
	r.Run(":8081")
}
