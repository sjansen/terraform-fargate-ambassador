package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type CheckHealth struct {
	FailAfter uint `help:"Report failure after N minutes."`
}

func (h *CheckHealth) Run(cfg *Config) error {
	resp, err := http.Get(cfg.AmbassadorURL + "/status")
	if err != nil {
		fmt.Println("error requesting status:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var status Status
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &status)
	if err != nil {
		fmt.Println("error decoding status:", err)
		os.Exit(1)
	}

	switch {
	case !status.Healthy:
		fmt.Println("unhealthy")
		os.Exit(1)
	case h.FailAfter > 0 && h.FailAfter*60 < status.UptimeSeconds:
		fmt.Println("fail-after triggered")
		os.Exit(1)
	}

	fmt.Println("healthy")
	return nil
}
