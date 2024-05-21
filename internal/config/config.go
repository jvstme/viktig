package config

import (
	"os"

	"viktig/internal/services/forwarder"
	"viktig/internal/services/http_server"
	"viktig/internal/services/http_server/handlers/metrics_handler"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ForwarderConfig  *forwarder.Config       `yaml:"forwarder_config" validate:"required"`
	HttpServerConfig *http_server.Config     `yaml:"http_config" validate:"required"`
	MetricsConfig    *metrics_handler.Config `yaml:"metrics_config"`
}

func LoadConfigFromFile(path string) (cfg *Config, err error) {
	if _, err = os.Stat(path); err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	if err = validator.New().Struct(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
