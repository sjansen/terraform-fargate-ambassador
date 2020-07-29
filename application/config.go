package main

import (
	"os"
)

type Config struct {
	Debug bool
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		Debug: os.Getenv("DEBUG") != "",
	}

	return cfg, nil
}
