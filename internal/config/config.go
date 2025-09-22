package config

import "os"

type Config struct {
	DatabaseDSN string
	HTTPPort    string
}

func Load() *Config {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://myuser:mypass@localhost:5432/tasksdb?sslmode=disable"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &Config{
		DatabaseDSN: dsn,
		HTTPPort:    port,
	}
}
