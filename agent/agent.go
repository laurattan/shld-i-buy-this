package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/laura838/shld-i-buy-this/models"
)

const (
	apiURL           = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
	modelID          = "claude-sonnet-4-6"
	maxTokens        = 1024
)

// contentBlock is the subset of fields we need from a response content block,
// regardless of its type (text, server_tool_use, web_search_tool_result, etc).
type contentBlock struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
	Text string `json:"text,omitempty"`
}

type apiRequest struct {
	Model     string                   `json:"model"`
	MaxTokens int                      `json:"max_tokens"`
	System    string                   `json:"system"`
	Messages  []map[string]any `json:"messages"`
	Tools     []map[string]any `json:"tools"`
}

type apiResponse struct {
	Content    []json.RawMessage `json:"content"`
	StopReason string            `json:"stop_reason"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Evaluate researches a product using Claude's web search tool and returns
// a structured buy/wait/find_alternative/do_not_buy decision.
func Evaluate(product string, profile models.UserProfile) (*models.Decision, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	systemPromptStr, err := buildSystemPrompt(profile)
	if err != nil {
		return nil, err
	}

	messages := []map[string]any{
		{"role": "user", "content": product},
	}

	tools := []map[string]any{
		{"type": "web_search_20260209", "name": "web_search"},
	}

	var finalText string

	for {
		reqBody := apiRequest{
			Model:     modelID,
			MaxTokens: maxTokens,
			System:    systemPromptStr,
			Messages:  messages,
			Tools:     tools,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}

		req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", anthropicVersion)

		callStart := time.Now()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("calling Anthropic API: %w", err)
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("API call took %s", time.Since(callStart))
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Anthropic API returned status %d: %s", resp.StatusCode, string(respBytes))
		}

		var apiResp apiResponse
		if err := json.Unmarshal(respBytes, &apiResp); err != nil {
			return nil, fmt.Errorf("unmarshaling response: %w", err)
		}
		if apiResp.Error != nil {
			return nil, fmt.Errorf("Anthropic API error: %s", apiResp.Error.Message)
		}

		var textParts []string
		for _, raw := range apiResp.Content {
			var block contentBlock
			if err := json.Unmarshal(raw, &block); err != nil {
				return nil, fmt.Errorf("unmarshaling content block: %w", err)
			}
			switch block.Type {
			case "server_tool_use", "tool_use":
				log.Println("tool call detected:", block.Name)
			case "text":
				textParts = append(textParts, block.Text)
			}
		}

		messages = append(messages, map[string]any{
			"role":    "assistant",
			"content": apiResp.Content,
		})

		// pause_turn means the model's turn was paused mid-way (e.g. a long
		// agentic turn) and should be continued by sending the conversation
		// back as-is. Any other stop reason means this is the final response.
		if apiResp.StopReason == "pause_turn" {
			continue
		}

		finalText = strings.Join(textParts, "")
		break
	}

	finalText = strings.TrimSpace(finalText)

	// The model is instructed to return only JSON, but in practice it may
	// wrap the object in markdown fences or add surrounding commentary.
	// Extract the outermost JSON object before parsing.
	start := strings.IndexByte(finalText, '{')
	end := strings.LastIndexByte(finalText, '}')
	if start == -1 || end == -1 || end < start {
		return nil, fmt.Errorf("no JSON object found in response: %s", finalText)
	}
	jsonText := finalText[start : end+1]

	var decision models.Decision
	if err := json.Unmarshal([]byte(jsonText), &decision); err != nil {
		return nil, fmt.Errorf("parsing decision JSON: %w (raw response: %s)", err, finalText)
	}

	switch decision.Label {
	case models.VerdictBuy, models.VerdictWait, models.VerdictFindAlternative, models.VerdictDoNotBuy:
	default:
		return nil, fmt.Errorf("invalid verdict label: %q", decision.Label)
	}

	return &decision, nil
}
