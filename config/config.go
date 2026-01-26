package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	dbURL := GetEnv("DATABASE_URL", "")
	if dbURL == "" {
		host := GetEnv("DB_HOST", "localhost")
		port := GetEnv("DB_PORT", "5432")
		user := GetEnv("DB_USER", "postgres")
		pass := GetEnv("DB_PASSWORD", "password")
		name := GetEnv("DB_NAME", "mini_exchange")
		ssl := GetEnv("DB_SSLMODE", "disable")
		dbURL = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + name + "?sslmode=" + ssl
	}

	return Config{
		DatabaseURL: dbURL,
		JWTSecret:   GetEnv("JWT_SECRET", "super-secret-dev-key"),
	}
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
