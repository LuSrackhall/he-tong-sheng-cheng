package config

import (
	"flag"
	"log"
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

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		log.Fatalf("FATAL: JWT_SECRET environment variable is required")
	}

	return cfg
}
