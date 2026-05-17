package config

import (
	"flag"
	"os"
)

type Config struct {
	Mode     string
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string
	JWTSecret string
	Port     string
}

func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Mode, "mode", "sqlite", "Database mode: sqlite or postgres")
	flag.StringVar(&cfg.DBHost, "db-host", "localhost", "Database host")
	flag.StringVar(&cfg.DBPort, "db-port", "5432", "Database port")
	flag.StringVar(&cfg.DBUser, "db-user", "postgres", "Database user")
	flag.StringVar(&cfg.DBPass, "db-pass", "", "Database password")
	flag.StringVar(&cfg.DBName, "db-name", "asset_leasing", "Database name")
	flag.StringVar(&cfg.Port, "port", "8080", "Server port")
	flag.Parse()

	cfg.JWTSecret = envDefault("JWT_SECRET", "asset-leasing-secret-change-me")

	return cfg
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
