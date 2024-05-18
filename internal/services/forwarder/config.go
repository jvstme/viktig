package forwarder

type Config struct {
	BotToken string `yaml:"bot_token" validate:"required"`
}
