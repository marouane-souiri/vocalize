package config

import "github.com/caarlos0/env/v11"

type Config struct {
	Discord struct {
		Token string `env:"TOKEN"`
	} `envPrefix:"DISCORD_"`
}

var Conf Config

// Load configuration
func Load() error {
	return env.Parse(&Conf)
}
