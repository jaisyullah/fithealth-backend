package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                   string
	DatabaseDSN            string
	RedisAddr              string
	SatusehatTokenURL      string
	SatusehatFHIRUrl       string
	SatusehatClientID      string
	SatusehatClientSecret  string
	LogLevel               string
	MaxWorkerPollInterval  int
}

func LoadFromEnv() *Config {
	port := getenv("PORT", "8080")
	dbHost := getenv("DB_HOST", "db")
	dbPort := getenv("DB_PORT", "5432")
	dbUser := getenv("DB_USER", "obsuser")
	dbPass := getenv("DB_PASS", "obssecret")
	dbName := getenv("DB_NAME", "obsdb")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)

	return &Config{
		Port:                  port,
		DatabaseDSN:           dsn,
		RedisAddr:             getenv("REDIS_ADDR", "redis:6379"),
		SatusehatTokenURL:     getenv("SATUSEHAT_TOKEN_URL", "https://sandbox.satusehat.example/oauth/token"),
		SatusehatFHIRUrl:      getenv("SATUSEHAT_FHIR_URL", "https://sandbox.satusehat.example/fhir"),
		SatusehatClientID:     getenv("SATUSEHAT_CLIENT_ID", ""),
		SatusehatClientSecret: getenv("SATUSEHAT_CLIENT_SECRET", ""),
		LogLevel:              getenv("LOG_LEVEL", "info"),
		MaxWorkerPollInterval: 1,
	}
}

func getenv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}
