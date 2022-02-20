package config

type Config struct {
	Persistent bool
}

func InitConfig() *Config {
	return &Config{
		Persistent: false,
	}
}
