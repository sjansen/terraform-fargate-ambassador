package main

import (
	"context"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
)

type Config struct {
	mutex sync.RWMutex

	env struct {
		RefreshURL string `envconfig:"AMBASSADOR_CONFIG_URL"`
	}
	refreshFailures int

	apiKey string

	RefreshDelay  time.Duration
	RefreshJitter time.Duration
}

func NewConfig(ctx context.Context) (*Config, error) {
	cfg := &Config{
		RefreshDelay:  1 * time.Hour,
		RefreshJitter: 5 * time.Minute,
	}
	if err := envconfig.Init(&cfg.env); err != nil {
		return nil, err
	}
	if err := cfg.fetchConfig(ctx); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *Config) APIKey() string {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()
	return cfg.apiKey
}

func (cfg *Config) Refresh(ctx context.Context, wg *sync.WaitGroup) {
	logger := zap.S()
	defer wg.Done()

	logger.Debug("Config refresher started.")
	for {
		delay := cfg.RefreshDelay + time.Duration(
			rand.Int63n(int64(cfg.RefreshJitter)),
		)
		select {
		case <-ctx.Done():
			logger.Debug("Config refresher stopped.")
			return
		case <-time.After(delay):
			logger.Debug("Refreshing config...")
			if err := cfg.fetchConfig(ctx); err != nil {
				logger.Debugw("Config refresh failed...",
					"error", err,
				)
				cfg.refreshFailures++
				// TODO report failed refresh
			} else {
				cfg.refreshFailures = 0
			}
		}
	}
}

func (cfg *Config) fetchConfig(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", cfg.env.RefreshURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	cfg.mutex.Lock()
	cfg.apiKey = string(body)
	cfg.mutex.Unlock()

	return nil
}
