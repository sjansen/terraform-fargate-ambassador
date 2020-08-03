package main

import (
	"errors"
	"os"
)

type Config struct {
	Debug    bool
	FillDisk bool

	AmbassadorURL  string
	ApplicationURL string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		Debug:    os.Getenv("DEBUG") != "",
		FillDisk: os.Getenv("FILL_DISK") != "",

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
