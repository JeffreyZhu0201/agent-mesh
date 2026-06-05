package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/gateway"
)

var configFile = flag.String("f", "api-gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var config gateway.GatewayConf
	conf.MustLoad(*configFile, &config)

	server := gateway.MustNewServer(config)
	defer server.Stop()

	fmt.Printf("Starting API Gateway at %d...\n", config.Port)
	server.Start()
}
