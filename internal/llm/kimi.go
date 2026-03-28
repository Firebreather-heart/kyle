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
	return SendPrompt(payload, c.LLMClient)
}

func (c *KIMIClient) GenerateComplex(messages []models.Prompt, tools []models.Tool) models.LLMResponse {
	payload := models.LLMRequest{
		Model:       "kimi-k2-thinking",
		Temperature: 0.3,
		Messages:    messages,
		Tools:       tools,
	}
	
	return SendPrompt(payload, c.LLMClient)
}

func NewKIMIClient(apiKey string) *KIMIClient {
	return &KIMIClient{
		LLMClient: models.LLMClient{
			APIKey: apiKey,
			HTTPClient: &http.Client{
				Timeout: 120 * time.Second,
			},
			RequestURI: "https://api.moonshot.cn/v1/chat/completions",
		},
	}
}