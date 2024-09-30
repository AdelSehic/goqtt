package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Logger *HttpLogger `json:"http-logger"`
	Connector *Connector `json:"connector"`
}

type HttpLogger struct {
	Url   string `json:"url"`
	Auth  string `json:"auth"`
}

type Connector struct {
	Port string `json:"port"`
}

func LoadConfig(file string) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
