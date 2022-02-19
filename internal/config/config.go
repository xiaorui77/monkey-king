package config

type Config struct {
	persistent bool
}

func InitConfig() *Config {
	return &Config{
		persistent: false,
	}
}
