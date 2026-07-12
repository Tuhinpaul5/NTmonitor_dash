package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	APP_ENV    string
	APP_DOMAIN string
	DBURL      string
	PORT       string
	JWT_SECRET string
	// SMTP Settings
	SMTP_HOST     string
	SMTP_PORT     string
	SMTP_USERNAME string
	SMTP_PASSWORD string
	SMTP_FROM     string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	// Set default PORT if not provided
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	return &Config{
		APP_ENV:       os.Getenv("APP_ENV"),
		APP_DOMAIN:    os.Getenv("APP_DOMAIN"),
		DBURL:         os.Getenv("DBURL"),
		PORT:          port,
		JWT_SECRET:    os.Getenv("JWT_SECRET"),
		SMTP_HOST:     os.Getenv("SMTP_HOST"),
		SMTP_PORT:     os.Getenv("SMTP_PORT"),
		SMTP_USERNAME: os.Getenv("SMTP_USERNAME"),
		SMTP_PASSWORD: os.Getenv("SMTP_PASSWORD"),
		SMTP_FROM:     os.Getenv("SMTP_FROM"),
	}
}
