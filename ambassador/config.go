package main

import (
	"errors"
	"os"
)

type Config struct {
	Debug   bool
	LinkURL string
	Queue   string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		Debug:   os.Getenv("DEBUG") != "",
		LinkURL: os.Getenv("APPLICATION"),
		Queue:   os.Getenv("QUEUE"),
	}

	var err error
	switch {
	case cfg.LinkURL == "":
		err = errors.New("Missing required setting: $APPLICATION")
	case cfg.Queue == "":
		err = errors.New("Missing required setting: $QUEUE")
	}

	return cfg, err
}
