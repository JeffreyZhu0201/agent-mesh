package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var configFile = flag.String("f", "plugin-svc.yaml", "the config file")

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

	r.GET("/api/plugins", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"plugins": []string{}})
	})

	fmt.Printf("Starting Plugin Service at port %d...\n", 8083)
	r.Run(":8083")
}
