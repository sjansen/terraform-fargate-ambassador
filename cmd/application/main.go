package main

import (
	"fmt"
	stdlog "log"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var cfg *Config
var startTime time.Time

var CLI struct {
	CheckHealth CheckHealth `cmd help:"Request status from application server."`
	Server      Server      `cmd help:"Run application in server mode."`
}

func main() {
	startTime = time.Now()

	config, err := NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	cfg = config

	if !cfg.Debug {
		log.Logger = log.Level(zerolog.InfoLevel)
	}
	stdlog.SetOutput(log.Logger)

	cli := kong.Parse(&CLI)
	err = cli.Run(cfg)
	cli.FatalIfErrorf(err)
}
