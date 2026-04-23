package llm

import (
	"net/http"
	"time"

	"github.com/firebreather-heart/kyle/internal/models"
)

type KIMIClient struct {
	models.LLMClient
}

func (c *KIMIClient) Generate(systemPrompt string, userPrompt string) models.LLMResponse {
	payload := models.LLMRequest{
		Model:       "kimi-k2.5",
		Temperature: 1,
		MaxTokens:   32000,
		Messages: []models.Prompt{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	return SendPrompt(payload, c.LLMClient)
}

func (c *KIMIClient) GenerateComplex(messages []models.Prompt, tools []models.Tool) models.LLMResponse {
	payload := models.LLMRequest{
		Model:       "kimi-k2.5",
		Temperature: 1,
		MaxTokens:   32000,
		Messages:    messages,
		Tools:       tools,
	}
	
	return SendPrompt(payload, c.LLMClient)
}

func (c *KIMIClient) UpdateAPIKey(apiKey string) {
	c.APIKey = apiKey
}

func NewKIMIClient(apiKey string) *KIMIClient {
	return &KIMIClient{
		LLMClient: models.LLMClient{
			APIKey: apiKey,
			HTTPClient: &http.Client{
				Timeout: 360 * time.Second,
			},
			RequestURI: "https://api.moonshot.ai/v1/chat/completions",
		},
	}
}