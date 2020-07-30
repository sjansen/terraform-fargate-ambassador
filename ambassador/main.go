package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger
var startTime time.Time

var CLI struct {
	CheckHealth CheckHealth `cmd help:"Request status from ambassador server."`
	Server      Server      `cmd help:"Run ambassador in server mode."`
}

func main() {
	startTime = time.Now()

	cfg, err := NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if cfg.Debug {
		logger = NewLogger(3)
	} else {
		logger = NewLogger(2)
	}

	cli := kong.Parse(&CLI)
	err = cli.Run(cfg)
	cli.FatalIfErrorf(err)
}
