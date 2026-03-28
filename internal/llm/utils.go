package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/firebreather-heart/kyle/internal/models"
)

func SendPrompt(
	systemInstruction string, 
	userPrompt string, 
	payload models.LLMRequest, 
	c models.LLMClient,
	) models.LLMResponse {

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return models.LLMResponse{
			Status:   "error",
			Response: "",
			Error:    err,
		}
	}
	req, err := http.NewRequest("POST", c.RequestURI, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.LLMResponse{
			Status:   "error",
			Response: "",
			Error:    err,
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return models.LLMResponse{
			Status:   "error",
			Response: "",
			Error:    err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return models.LLMResponse{
			Status:   "error",
			Response: "",
			Error:    fmt.Errorf("API returned status: %d: %s", resp.StatusCode, string(bodyBytes)),
		}
	}

	var apiData models.RawAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return models.LLMResponse{Status: "error", Error: err}
	}
	if len(apiData.Choices) == 0 {
		return models.LLMResponse{
			Status:   "error",
			Response: "",
			Error:    fmt.Errorf("API returned empty choices"),
		}
	}

	return models.LLMResponse{
		Status:   "success",
		Response: apiData.Choices[0].Message.Content,
		Error:    nil,
	}
}
