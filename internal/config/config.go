package config

import (
	"log"
	"os"
	"strings"
	"github.com/joho/godotenv"
)

type AppConfig struct{
	GEMINI_API_KEYS []string
	KIMI_API_KEYS   []string
	ServerPort string
	RedisAddr  string
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
	geminiKeysCSV := getEnvOrDefault("GEMINI_API_KEYS", getEnvOrDefault("GEMINI_API_KEY", ""))
	var geminiKeys []string
	if geminiKeysCSV != "" {
		for _, k := range strings.Split(geminiKeysCSV, ",") {
			if trimmed := strings.TrimSpace(k); trimmed != "" {
				geminiKeys = append(geminiKeys, trimmed)
			}
		}
	}

	kimiKeysCSV := getEnvOrDefault("KIMI_API_KEYS", getEnvOrDefault("KIMI_API_KEY", ""))
	var kimiKeys []string
	if kimiKeysCSV != "" {
		for _, k := range strings.Split(kimiKeysCSV, ",") {
			if trimmed := strings.TrimSpace(k); trimmed != "" {
				kimiKeys = append(kimiKeys, trimmed)
			}
		}
	}

	return &AppConfig{
		GEMINI_API_KEYS:     geminiKeys,
		KIMI_API_KEYS:       kimiKeys,
		ServerPort:          getEnvOrDefault("SERVER_PORT", "8080"),
		RedisAddr:           getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		MaxScrapeTokens:     10000,
		DAILY_RATE_LIMIT:    5,
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: os.Getenv("CLOUDINARY_API_SECRET"),
	}
}

func (c *AppConfig) GetProviderKeys(provider string) []string {
	switch provider {
	case "gemini":
		return c.GEMINI_API_KEYS
	case "kimi":
		return c.KIMI_API_KEYS
	}
	return nil
}

func getEnvOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != ""{
		return value
	}
	log.Printf("WARNING: %s not set, using fallback: %s", key, fallback)
	return fallback
}