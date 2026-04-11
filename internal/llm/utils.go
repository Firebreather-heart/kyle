package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/firebreather-heart/kyle/internal/models"
)

func SendPrompt(payload models.LLMRequest, c models.LLMClient) models.LLMResponse {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return models.LLMResponse{Error: err}
	}

	req, err := http.NewRequest("POST", c.RequestURI, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.LLMResponse{Error: err}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return models.LLMResponse{Error: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return models.LLMResponse{
			Error: fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes)),
		}
	}

	var apiData models.RawAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return models.LLMResponse{Error: err}
	}

	if len(apiData.Choices) == 0 {
		return models.LLMResponse{Error: fmt.Errorf("no choices in response")}
	}

	choice := apiData.Choices[0].Message

	if len(choice.ToolCalls) > 0 {
		return models.LLMResponse{
			ReasoningContent: choice.ReasoningContent,
			ToolCall:         &choice.ToolCalls[0],
			Response:         choice.Content,
		}
	}

	return models.LLMResponse{
		ReasoningContent: choice.ReasoningContent,
		Response:         choice.Content,
	}
}
