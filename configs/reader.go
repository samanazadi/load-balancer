package configs

import (
	"encoding/json"
	"github.com/samanazadi/load-balancer/internal/logging"
	"os"
)

const (
	configPath   = "/etc/load-balancer/config.json"
	strategyPath = "/etc/load-balancer/strategy.json"
)

type Config struct {
	Nodes       []string `json:"nodes"`
	HealthCheck struct {
		Active struct {
			MaxRetry   int `json:"maxRetry"`
			RetryDelay int `json:"retryDelay"`
			MaxAttempt int `json:"maxAttempt"`
		} `json:"active"`
		Passive struct {
			Period  int `json:"period"`
			Timeout int `json:"timeout"`
		} `json:"passive"`
	} `json:"healthCheck"`
	Strategy struct {
		Name   string `json:"name"`
		Params map[string]any
	} `json:"strategy"`
}

func Read() Config {
	var config Config

	// read config.json file
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		logging.Logger.Fatalf("Cannot read config file: %s", err.Error())
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		logging.Logger.Fatalf("Cannot unmarshal config file: %s", err.Error())
	}

	// read strategy.json file
	var strategyParams map[string]any
	strategyBytes, err := os.ReadFile(strategyPath)
	if err != nil {
		logging.Logger.Fatalf("Cannot read strategy file: %s", err.Error())
	}
	err = json.Unmarshal(strategyBytes, &strategyParams)
	if err != nil {
		logging.Logger.Fatalf("Cannot unmarshal strategy file: %s", err.Error())
	}
	config.Strategy.Params = strategyParams

	return config
}
