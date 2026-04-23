package models

import (
	"net/http"
)

type TokenRequest struct {
	Fingerprint string `json:"fingerprint"`
}

type ClientRequest struct {
	Topic    string `json:"topic"`
	Provider string `json:"provider"`
	Format   string `json:"format"`
}

type SSEUpdate struct {
	Status   string   `json:"status,omitempty"`
	Progress string   `json:"progress,omitempty"`
	NewLogs  []string `json:"newLogs,omitempty"`
	Complete bool     `json:"complete,omitempty"`
	Result   string   `json:"result,omitempty"`
}

type LLMClient struct {
	APIKey     string
	HTTPClient *http.Client
	RequestURI string
	ModelName  string
}

type Prompt struct {
	Role             string     `json:"role"`
	Content          string     `json:"content"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string     `json:"tool_call_id,omitempty"`
}

type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type LLMRequest struct {
	Model       string   `json:"model"`
	Temperature float64  `json:"temperature"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Messages    []Prompt `json:"messages"`
	Tools       []Tool   `json:"tools,omitempty"`
}

type LLMResponse struct {
	Response         string
	ReasoningContent string
	ToolCall         *ToolCall
	StatusCode       int
	Error            error
}

type ToolCall struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type RawAPIResponse struct {
	Choices []struct {
		Message struct {
			Role             string     `json:"role"`
			Content          string     `json:"content"`
			ReasoningContent string     `json:"reasoning_content,omitempty"`
			ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
}


type ContentBlock struct {
	Type            string     `json:"type"`
	Content         string     `json:"content,omitempty"`
	BackgroundColor string     `json:"background_color,omitempty"`
	TextColor       string     `json:"text_color,omitempty"`
	Icon            string     `json:"icon,omitempty"`
	Language        string     `json:"language,omitempty"`
	Headers         []string   `json:"headers,omitempty"`
	Rows            [][]string `json:"rows,omitempty"`
	Items           []string   `json:"items,omitempty"`
}

type AIBlock struct {
	Type            string     `json:"type"`
	Content         string     `json:"content,omitempty"`
	HeadingFont     string     `json:"heading_font,omitempty"`
	BodyFont        string     `json:"body_font,omitempty"`
	PrimaryColor    string     `json:"primary_color,omitempty"`
	SecondaryColor  string     `json:"secondary_color,omitempty"`
	LayoutDensity   string     `json:"layout_density,omitempty"`
	BackgroundColor string     `json:"background_color,omitempty"`
	TextColor       string     `json:"text_color,omitempty"`
	Icon            string     `json:"icon,omitempty"`
	Language        string     `json:"language,omitempty"`
	Headers         []string   `json:"headers,omitempty"`
	Rows            [][]string `json:"rows,omitempty"`
	Items           []string   `json:"items,omitempty"`
}