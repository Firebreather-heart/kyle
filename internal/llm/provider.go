package llm

import (
	"fmt"

	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/models"
)

type Provider interface{
	Generate(systemPrompt string, userPrompt string) models.LLMResponse
	GenerateComplex(messages []models.Prompt, tools []models.Tool) models.LLMResponse
}

func NewProvider(cfg *config.AppConfig, providerType string) (Provider, error) {
	switch providerType {
	case "gemini":
		return NewGeminiClient(cfg.GEMINI_API_KEY), nil
	case "kimi":
		return NewKIMIClient(cfg.KIMI_API_KEY), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}
