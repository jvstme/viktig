package forwarder

type Config struct {
	VkConfig *VkConfig `yaml:"vk_config"`
	TgConfig *TgConfig `yaml:"tg_config" validate:"required"`
}

type VkConfig struct {
}

type TgConfig struct {
	Token  string `yaml:"token" validate:"required"`
	ChatId int    `yaml:"chat_id" validate:"required"`
}
