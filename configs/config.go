package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

type Algorithm struct {
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
	Algorithm   Algorithm   `json:"algorithm"`
	Checker     Checker     `json:"checker"`
}

func New(cfgPath string) (*Config, error) {
	var config Config
	var (
		configPath    = filepath.Join(cfgPath, "config.json")
		algorithmPath = filepath.Join(cfgPath, "algorithm.json")
		checkerPath   = filepath.Join(cfgPath, "checker.json")
	)

	// read config.json file
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %s", err.Error())
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal config file: %s", err.Error())
	}

	// read algorithm.json file
	params, err := readConfigFile(algorithmPath, "algorithm")
	if err != nil {
		return nil, err
	}
	config.Algorithm.Params = params

	// read checker.json file
	params, err = readConfigFile(checkerPath, "checker")
	if err != nil {
		return nil, err
	}
	config.Checker.Params = params

	return &config, nil
}

func readConfigFile(path string, configType string) (map[string]any, error) {
	var algorithmParams map[string]any
	algorithmBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s file: %s", configType, err.Error())
	}
	err = json.Unmarshal(algorithmBytes, &algorithmParams)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal %s file: %s", configType, err.Error())
	}
	return algorithmParams, nil
}
