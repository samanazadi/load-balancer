package configs

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	configPath   = "/etc/load-balancer/config.json"
	strategyPath = "/etc/load-balancer/strategy.json"
	checkerPath  = "/etc/load-balancer/checker.json"
)

type ActiveHealthCheck struct {
	MaxRetry   int `json:"maxRetry"`
	RetryDelay int `json:"retryDelay"`
}

type PassiveHealthCheck struct {
	Period  int `json:"period"`
	Timeout int `json:"timeout"`
}

type HealthCheck struct {
	Active  ActiveHealthCheck  `json:"active"`
	Passive PassiveHealthCheck `json:"passive"`
}

type Strategy struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"-,"`
}

type Checker struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"-,"`
}

type Config struct {
	Port        int         `json:"port"`
	Nodes       []string    `json:"nodes"`
	HealthCheck HealthCheck `json:"healthCheck"`
	Strategy    Strategy    `json:"strategy"`
	Checker     Checker     `json:"checker"`
}

func New() (*Config, error) {
	var config Config

	// read config.json file
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %s", err.Error())
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal config file: %s", err.Error())
	}

	// read strategy.json file
	params, err := readConfigFile(strategyPath, "strategy")
	if err != nil {
		return nil, err
	}
	config.Strategy.Params = params

	// read checker.json file
	params, err = readConfigFile(checkerPath, "checker")
	if err != nil {
		return nil, err
	}
	config.Checker.Params = params

	return &config, nil
}

func readConfigFile(path string, configType string) (map[string]any, error) {
	var strategyParams map[string]any
	strategyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s file: %s", configType, err.Error())
	}
	err = json.Unmarshal(strategyBytes, &strategyParams)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal %s file: %s", configType, err.Error())
	}
	return strategyParams, nil
}
