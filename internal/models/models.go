package models

import (
	"net/http"
)


type Prompt struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

type LLMRequest struct {
	Model string `json:"model"`
	Temperature float64 `json:"temperature"`
	Messages []Prompt `json:"messages"`
}

type LLMResponse struct {
	Status string `json:"status"`
	Response string `json:"response"`
	Error error `json:"error"`
}

type RawAPIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type LLMClient struct {
	APIKey     string
	HTTPClient *http.Client
	RequestURI string
}

type ClientRequest struct {
	Topic string `json:"topic"`
	Provider string `json:"provider"`
}
