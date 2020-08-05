package main

import "github.com/vrischmann/envconfig"

type Config struct {
	ReadyURL string `envconfig:"InitializationCallbackUrl"`

	LogLevel      int `envconfig:"Runner__LogLevel,default=2"`
	QueueCapacity int `envconfig:"optional"`
	ReadyDelayMs  int `envconfig:"InitializationRetryDelayMs,default=1000"`
	ReadyRetries  int `envconfig:"InitializationAttempts,default=5"`
	WorkerCount   int `envconfig:"Runner__CountThreads,default=1"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Init(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
