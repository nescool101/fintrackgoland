package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APIKey       string
	DatabaseURL  string
	EmailHost    string
	EmailPort    int
	EmailUser    string
	EmailPass    string
	Recipient    string
	RunCron      bool
	AuthUsername string
	AuthPassword string
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
		APIKey:      os.Getenv("API_KEY"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		EmailHost:   os.Getenv("EMAIL_HOST"),
		EmailPort:   emailPort,
		EmailUser:   os.Getenv("EMAIL_USER"),
		EmailPass:   os.Getenv("EMAIL_PASS"),
		Recipient:   os.Getenv("EMAIL_RECIPIENT"),
		RunCron:     os.Getenv("RUN_CRON") == "true",
	}

	// Basic validation
	if config.APIKey == "" || config.DatabaseURL == "" || config.EmailHost == "" ||
		config.EmailUser == "" || config.EmailPass == "" || config.Recipient == "" ||
		config.AuthUsername == "" || config.AuthPassword == "" {
		log.Fatal("Missing required environment variables")
	}

	return config
}
