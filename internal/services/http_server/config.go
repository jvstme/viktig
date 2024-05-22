package http_server

type Config struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}
