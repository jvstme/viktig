package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	TgBotToken       string             `yaml:"tg_bot_token" validate:"required"`
	VkApiToken       string             `yaml:"vk_api_token" validate:"required"`
	MetricsAuthToken string             `yaml:"metrics_auth_token"`
	Communities      []*CommunityConfig `yaml:"communities" validate:"required,dive"`
}

type CommunityConfig struct {
	HookId             string `yaml:"hook_id" validate:"required"`
	SecretKey          string `yaml:"secret_key" validate:"required"`
	ConfirmationString string `yaml:"confirmation_string" validate:"required"`
	TgChatId           int    `yaml:"tg_chat_id" validate:"required"`
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
