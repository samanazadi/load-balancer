package configs

import (
	"encoding/json"
	"github.com/samanazadi/load-balancer/internal/logging"
	"os"
)

const (
	configPath   = "/etc/load-balancer/config.json"
	strategyPath = "/etc/load-balancer/strategy.json"
	checkerPath  = "/etc/load-balancer/checker.json"
)

type config struct {
	Port        int      `json:"port"`
	Nodes       []string `json:"nodes"`
	HealthCheck struct {
		Active struct {
			MaxRetry   int `json:"maxRetry"`
			RetryDelay int `json:"retryDelay"`
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
	Checker struct {
		Name   string `json:"name"`
		Params map[string]any
	} `json:"checker"`
}

var Config config

func read() config {
	var config config

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

	// read checker.json file
	var checkerParams map[string]any
	checkerBytes, err := os.ReadFile(checkerPath)
	if err != nil {
		logging.Logger.Fatalf("Cannot read checker file: %s", err.Error())
	}
	err = json.Unmarshal(checkerBytes, &checkerParams)
	if err != nil {
		logging.Logger.Fatalf("Cannot unmarshal checker file: %s", err.Error())
	}
	config.Checker.Params = checkerParams

	return config
}

func init() {
	Config = read()
}
