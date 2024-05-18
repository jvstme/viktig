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
	TempConfig       *TempConfig             `yaml:"temp_config" validate:"required"`
	MetricsConfig    *metrics_handler.Config `yaml:"temp_config"`
}

type TempConfig struct {
	TgChatId           int    `yaml:"tg_chat_id" validate:"required"`
	HookId             string `yaml:"community_hook_id" validate:"required"`
	ConfirmationString string `yaml:"vk_confirmation_string" validate:"required"`
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

	//if err = defaults.Set(cfg); err != nil {
	//	return nil, err
	//}

	if err = validator.New().Struct(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
