package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	Mode           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPass         string
	DBName         string
	JWTSecret      string
	Port           string
	UploadDir      string
	DefaultCurrency string
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() *Config {
	flagMode := flag.String("mode", "sqlite", "Database mode: sqlite or postgres")
	flagDBHost := flag.String("db-host", "localhost", "Database host")
	flagDBPort := flag.String("db-port", "5432", "Database port")
	flagDBUser := flag.String("db-user", "postgres", "Database user")
	flagDBPass := flag.String("db-pass", "", "Database password")
	flagDBName := flag.String("db-name", "design_platform", "Database name")
	flagPort := flag.String("port", "8080", "Server port")
	flag.Parse()

	cfg := &Config{
		Mode:           envOrDefault("MODE", *flagMode),
		DBHost:         envOrDefault("DB_HOST", *flagDBHost),
		DBPort:         envOrDefault("DB_PORT", *flagDBPort),
		DBUser:         envOrDefault("DB_USER", *flagDBUser),
		DBPass:         envOrDefault("DB_PASS", *flagDBPass),
		DBName:         envOrDefault("DB_NAME", *flagDBName),
		Port:           envOrDefault("PORT", *flagPort),
		UploadDir:      envOrDefault("UPLOAD_DIR", "./uploads"),
		DefaultCurrency: envOrDefault("DEFAULT_CURRENCY", "CNY"),
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		log.Fatalf("FATAL: JWT_SECRET environment variable is required")
	}

	return cfg
}
