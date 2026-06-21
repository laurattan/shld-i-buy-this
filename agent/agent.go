package agent

import (
	"bytes"
	"context"
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

type contentBlock struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
	Text string `json:"text,omitempty"`
}

type apiRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	System    string           `json:"system"`
	Messages  []map[string]any `json:"messages"`
	Tools     []map[string]any `json:"tools,omitempty"`
}

type apiResponse struct {
	Content    []json.RawMessage `json:"content"`
	StopReason string            `json:"stop_reason"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// callClaude runs the agentic tool-use loop and returns the final text response.
// If withTools is true, web_search (max 5 uses) and web_fetch (max 3 uses) are attached.
func callClaude(ctx context.Context, systemPrompt string, messages []map[string]any, withTools bool) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	var tools []map[string]any
	if withTools {
		tools = []map[string]any{
			{"type": "web_search_20260209", "name": "web_search", "max_uses": 4},
			{"type": "web_fetch_20260209", "name": "web_fetch", "max_uses": 4},
		}
	}

	if len(messages) > 0 {
		if query, ok := messages[0]["content"].(string); ok {
			log.Printf("query: %s", query)
		}
	}

	var finalText string

	for {
		reqBody := apiRequest{
			Model:     modelID,
			MaxTokens: maxTokens,
			System:    systemPrompt,
			Messages:  messages,
			Tools:     tools,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return "", fmt.Errorf("marshaling request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return "", fmt.Errorf("building request: %w", err)
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", anthropicVersion)

		callStart := time.Now()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("calling Anthropic API: %w", err)
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("API call took %s", time.Since(callStart))
		if err != nil {
			return "", fmt.Errorf("reading response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("Anthropic API returned status %d: %s", resp.StatusCode, string(respBytes))
		}

		var apiResp apiResponse
		if err := json.Unmarshal(respBytes, &apiResp); err != nil {
			return "", fmt.Errorf("unmarshaling response: %w", err)
		}
		if apiResp.Error != nil {
			return "", fmt.Errorf("Anthropic API error: %s", apiResp.Error.Message)
		}

		var textParts []string
		for _, raw := range apiResp.Content {
			var block contentBlock
			if err := json.Unmarshal(raw, &block); err != nil {
				return "", fmt.Errorf("unmarshaling content block: %w", err)
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

		if apiResp.StopReason == "pause_turn" {
			continue
		}

		finalText = strings.Join(textParts, "")
		break
	}

	result := strings.TrimSpace(finalText)
	log.Printf("output: %s", result)
	return result, nil
}

// extractJSON extracts the outermost JSON object from text that may be wrapped
// in markdown fences or surrounded by commentary.
func extractJSON(text string) (string, error) {
	start := strings.IndexByte(text, '{')
	end := strings.LastIndexByte(text, '}')
	if start == -1 || end == -1 || end < start {
		return "", fmt.Errorf("no JSON object found in response: %s", text)
	}
	return text[start : end+1], nil
}

// Evaluate researches a product and returns a structured buy decision.
func Evaluate(product string, profile models.UserProfile) (*models.Decision, error) {
	ctx := context.Background()

	analysis, err := research(ctx, product, profile)
	if err != nil {
		return nil, fmt.Errorf("research step: %w", err)
	}

	return &models.Decision{
		ProductName:  analysis.ProductName,
		Label:        analysis.Verdict,
		Reasoning:    analysis.Reasoning,
		Alternatives: analysis.Alternatives,
	}, nil
}
