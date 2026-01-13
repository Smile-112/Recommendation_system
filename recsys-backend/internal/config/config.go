package config

import "os"

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

type Config struct {
	Addr string
	DB   DBConfig
}

func FromEnv() Config {
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{
		Addr: addr,
		DB: DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Name:     os.Getenv("DB_NAME"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
		},
	}
}
