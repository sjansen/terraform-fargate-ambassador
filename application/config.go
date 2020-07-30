package main

import (
	"errors"
	"os"
)

type Config struct {
	Debug bool

	AmbassadorURL  string
	ApplicationURL string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		Debug: os.Getenv("DEBUG") != "",

		AmbassadorURL:  os.Getenv("AMBASSADOR"),
		ApplicationURL: os.Getenv("APPLICATION"),
	}

	var err error
	switch {
	case cfg.AmbassadorURL == "":
		err = errors.New("Missing required setting: $AMBASSADOR")
	case cfg.ApplicationURL == "":
		err = errors.New("Missing required setting: $APPLICATION")
	}

	return cfg, err
}
