package config

import "os"

type Config struct {
	ClientToken string
}

// Load configuration
func Load() *Config {
	return &Config{
		ClientToken: loadEnv("CLIENT_TOKEN", "CLIENT TOKEN GOES HERE"),
	}
}

func loadEnv(key string, defaultVal string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return val
}
