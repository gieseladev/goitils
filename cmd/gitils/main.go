package main

import (
	"flag"
	"log"

	"github.com/gieseladev/goitils/pkg/gitils"
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	configFlag "github.com/micro/go-config/source/flag"
)

func getConfigLocationFromCmdArgs() string {
	configLocationPtr := flag.String("config", "config.json", "specify the location of the config file")

	flag.Parse()
	return *configLocationPtr
}

func main() {
	configLocation := getConfigLocationFromCmdArgs()
	log.Println("config location:", configLocation)

	rawConf := config.NewConfig()
	rawConf.Load(
		env.NewSource(),
		configFlag.NewSource(),
		file.NewSource(file.WithPath(configLocation)),
	)

	conf := gitils.NewConfig()
	rawConf.Scan(&conf)

	log.Fatal(gitils.Start(conf))
}
