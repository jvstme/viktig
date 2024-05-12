package http_server

type Config struct {
	Address            string `yaml:"address"`
	Port               int    `yaml:"port"`
	MetricsAuthToken   string `yaml:"metrics_auth_token"`
	ConfirmationString string `yaml:"confirmation_string"`
}
