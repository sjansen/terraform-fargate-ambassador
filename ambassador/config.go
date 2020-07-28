package main

import (
	"errors"
	"os"
)

type Config struct {
	Debug bool
	Queue string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		Debug: os.Getenv("DEBUG") != "",
		Queue: os.Getenv("QUEUE"),
	}

	var err error
	switch {
	case cfg.Queue == "":
		err = errors.New("Missing required setting: $QUEUE")
	}

	return cfg, err
}
