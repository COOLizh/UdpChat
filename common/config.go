package common

import (
	"flag"

	"github.com/BurntSushi/toml"
)

//Config : configuraton of server
type Config struct {
	BindAddr string
	Network  string
}

//NewConfig : creates new configuration of server
func NewConfig() *Config {
	return &Config{
		BindAddr: "127.0.0.1:8080",
		Network:  "udp4",
	}
}

//GetConfig : returns config
func GetConfig() *Config {
	var configPath string
	flag.StringVar(&configPath, "config-path", "../configs/server.toml", "config file path")
	flag.Parse()
	config := NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	HandleError(err, ErrorFatal)
	return config
}
