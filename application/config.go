package main

import (
	"errors"
	"os"
)

type Config struct {
	Debug   bool
	LinkURL string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		LinkURL: os.Getenv("AMBASSADOR"),
		Debug:   os.Getenv("DEBUG") != "",
	}

	var err error
	switch {
	case cfg.LinkURL == "":
		err = errors.New("Missing required setting: $AMBASSADOR")
	}

	return cfg, err
}
