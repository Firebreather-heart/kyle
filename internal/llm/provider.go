package llm

import (
	"fmt"

	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/models"
)

type Provider interface{
	Generate(systemPrompt string, userPrompt string) models.LLMResponse
	GenerateComplex(messages []models.Prompt, tools []models.Tool) models.LLMResponse
	UpdateAPIKey(apiKey string)
}

func NewProvider(cfg *config.AppConfig, providerType string) (Provider, error) {
	switch providerType {
	case "gemini":
		key := ""
		if len(cfg.GEMINI_API_KEYS) > 0 {
			key = cfg.GEMINI_API_KEYS[0]
		}
		return NewGeminiClient(key), nil
	case "kimi":
		key := ""
		if len(cfg.KIMI_API_KEYS) > 0 {
			key = cfg.KIMI_API_KEYS[0]
		}
		return NewKIMIClient(key), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}
