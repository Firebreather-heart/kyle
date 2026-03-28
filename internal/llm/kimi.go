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
		Model:       "kimi-k2-thinking",
		Temperature: 0.2,
		Messages: []models.Prompt{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	return SendPrompt(systemPrompt, userPrompt, payload, c.LLMClient)
}

func NewKimiClient(apiKey string) *KIMIClient {
	return &KIMIClient{
		LLMClient: models.LLMClient{
			APIKey: apiKey,
			HTTPClient: &http.Client{
				Timeout: 30 * time.Second,
			},
			RequestURI: "https://api.moonshot.cn/v1/chat/completions",
		},
	}
}