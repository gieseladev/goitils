package main

import (
	"flag"
	"os"

	"github.com/gieseladev/goitils/pkg/gitils"
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	configFlag "github.com/micro/go-config/source/flag"
	log "github.com/sirupsen/logrus"
)

func getConfigLocationFromCmdArgs() string {
	configLocationPtr := flag.String("config", "config.json", "specify the location of the config file")

	flag.Parse()
	return *configLocationPtr
}

func setupLogging() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

}

func main() {
	setupLogging()

	configLocation := getConfigLocationFromCmdArgs()
	log.Debugf("config location: %q", configLocation)

	rawConf := config.NewConfig()
	log.Trace("loading config")
	rawConf.Load(
		env.NewSource(),
		configFlag.NewSource(),
		file.NewSource(file.WithPath(configLocation)),
	)

	conf := gitils.NewConfig()
	rawConf.Scan(&conf)

	log.Fatal(gitils.Start(conf))
}
