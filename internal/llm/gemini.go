package llm 

import (
	"time"
	"net/http"
	"github.com/firebreather-heart/kyle/internal/models"
)

type GeminiClient struct {
	models.LLMClient
}

func (c *GeminiClient) Generate(systemPrompt string, userPrompt string) models.LLMResponse {
	payload := models.LLMRequest{
		Model:       "gemini-3.1-flash-preview",
		Temperature: 0.2,
		Messages: []models.Prompt{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	return SendPrompt(systemPrompt, userPrompt, payload, c.LLMClient)
}

func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		LLMClient: models.LLMClient{
			APIKey: apiKey,
			HTTPClient: &http.Client{
				Timeout: 30 * time.Second,
			},
			RequestURI: "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
		},
	}
}