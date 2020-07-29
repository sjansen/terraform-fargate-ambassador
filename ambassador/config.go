package main

import (
	"errors"
	"os"
)

type Config struct {
	Debug  bool
	AppURL string
	Queue  string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		Debug:  os.Getenv("DEBUG") != "",
		AppURL: os.Getenv("APPURL"),
		Queue:  os.Getenv("QUEUE"),
	}

	var err error
	switch {
	case cfg.AppURL == "":
		err = errors.New("Missing required setting: $APPURL")
	case cfg.Queue == "":
		err = errors.New("Missing required setting: $QUEUE")
	}

	return cfg, err
}
