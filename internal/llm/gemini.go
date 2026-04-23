package llm

import (
	"net/http"
	"time"

	"github.com/firebreather-heart/kyle/internal/models"
)

type GeminiClient struct {
	models.LLMClient
}

func (c *GeminiClient) Generate(systemPrompt string, userPrompt string) models.LLMResponse {
	payload := models.LLMRequest{
		Model:       "gemini-2.5-flash",
		Temperature: 0.4,
		MaxTokens:   32000,
		Messages: []models.Prompt{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	return SendPrompt(payload, c.LLMClient)
}

func (c *GeminiClient) GenerateComplex(messages []models.Prompt, tools []models.Tool) models.LLMResponse {
	payload := models.LLMRequest{
		Model:       "gemini-2.5-flash",
		Temperature: 0.3,
		MaxTokens:   32000,
		Messages:    messages,
		Tools:       tools,
	}
	
	return SendPrompt(payload, c.LLMClient)
}

func (c *GeminiClient) UpdateAPIKey(apiKey string) {
	c.APIKey = apiKey
}

func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		LLMClient: models.LLMClient{
			APIKey: apiKey,
			HTTPClient: &http.Client{
				Timeout: 360 * time.Second,
			},
			RequestURI: "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
		},
	}
}