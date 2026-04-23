package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type AppConfig struct{
	GEMINI_API_KEY string
	KIMI_API_KEY string
	ServerPort string
	MaxScrapeTokens int
	DAILY_RATE_LIMIT int
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
}

/*
  Load system configuration for app
*/
func Load() *AppConfig{
	err  := godotenv.Load()
	if err != nil {
		log.Println("NO .env found, falling back to environment")
	}
	return &AppConfig{
		GEMINI_API_KEY: getEnvOrDefault("GEMINI_API_KEY", ""),
		KIMI_API_KEY: getEnvOrDefault("KIMI_API_KEY", ""),
		ServerPort: getEnvOrDefault("SERVER_PORT", "8080"),
		MaxScrapeTokens: 10000,
		DAILY_RATE_LIMIT: 5,
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: os.Getenv("CLOUDINARY_API_SECRET"),
	}
}

func getEnvOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != ""{
		return value
	}
	log.Printf("WARNING: %s not set, using fallback: %s", key, fallback)
	return fallback
}