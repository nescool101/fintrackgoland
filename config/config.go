package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	FMPAPIKey       string // Financial Modeling Prep API Key (primary)
	AlphaVantageKey string // Alpha Vantage API Key (for indices)
	DatabaseURL     string
	EmailHost       string
	EmailPort       int
	EmailUser       string
	EmailPass       string
	Recipient       string
	RunCron         bool
	AuthUsername    string
	AuthPassword    string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Parse EMAIL_PORT as integer
	emailPortStr := os.Getenv("EMAIL_PORT")
	emailPort, err := strconv.Atoi(emailPortStr)
	if err != nil {
		log.Printf("Invalid EMAIL_PORT '%s', defaulting to 587", emailPortStr)
		emailPort = 587
	}

	config := &Config{
		FMPAPIKey:       os.Getenv("FMP_API_KEY"),       // Financial Modeling Prep API Key
		AlphaVantageKey: os.Getenv("ALPHA_VANTAGE_KEY"), // Alpha Vantage API Key
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		EmailHost:       os.Getenv("EMAIL_HOST"),
		EmailPort:       emailPort,
		EmailUser:       os.Getenv("EMAIL_USER"),
		EmailPass:       os.Getenv("EMAIL_PASS"),
		Recipient:       os.Getenv("EMAIL_RECIPIENT"),
		RunCron:         os.Getenv("RUN_CRON") == "true",
		AuthUsername:    os.Getenv("AUTH_USERNAME"),
		AuthPassword:    os.Getenv("AUTH_PASSWORD"),
	}

	// Basic validation
	if config.FMPAPIKey == "" || config.AlphaVantageKey == "" || config.DatabaseURL == "" ||
		config.EmailHost == "" || config.EmailUser == "" || config.EmailPass == "" ||
		config.Recipient == "" || config.AuthUsername == "" || config.AuthPassword == "" {
		log.Fatal("Missing required environment variables")
	}

	return config
}
