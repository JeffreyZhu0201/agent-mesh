package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "user-svc.yaml", "the config file")

type Config struct {
	Name  string `yaml:"name"`
	Port  int    `yaml:"port"`
	MySQL struct {
		Host            string `yaml:"host"`
		Port            int    `yaml:"port"`
		User            string `yaml:"user"`
		Password        string `yaml:"password"`
		Database        string `yaml:"database"`
		MaxOpenConns    int    `yaml:"maxOpenConns"`
		MaxIdleConns    int    `yaml:"maxIdleConns"`
		ConnMaxLifetime int    `yaml:"connMaxLifetime"`
	} `yaml:"mysql"`
	JWT struct {
		Secret  string `yaml:"secret"`
		SignKey string `yaml:"signKey"`
		Expiry int    `yaml:"expiry"`
	} `yaml:"jwt"`
}

func main() {
	flag.Parse()

	var config Config
	conf.MustLoad(*configFile, &config)

	server := zrpc.NewServer(fmt.Sprintf(":%d", config.Port))
	defer server.Stop()

	fmt.Printf("Starting User Service at port %d...\n", config.Port)
	server.Start()
}
