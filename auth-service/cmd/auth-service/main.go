// main ...
package main

import (
	"auth/internal/app/server"
	"flag"
	"log"

	"github.com/BurntSushi/toml"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "config/config.toml", "path to config file")

}

func main() {
	cfg := server.NewConfig()
	_, err := toml.DecodeFile(configPath, cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Start(cfg); err != nil {
		log.Fatal(err)
	}

}
