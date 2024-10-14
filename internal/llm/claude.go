package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const ClaudeAPIURL = "https://api.anthropic.com/v1/messages"

// ClaudeClient implémente l'interface LLMClient pour le modèle Claude d'Anthropic
type ClaudeClient struct {
	APIKey      string
	Model       string
	ContextSize int
	Timeout     time.Duration
	HTTPClient  *http.Client
}

// NewClaudeClient crée et retourne une nouvelle instance de ClaudeClient
func NewClaudeClient(apiKey, model string, contextSize int, timeout time.Duration) *ClaudeClient {
	return &ClaudeClient{
		APIKey:      apiKey,
		Model:       model,
		ContextSize: contextSize,
		Timeout:     timeout,
		HTTPClient:  &http.Client{Timeout: timeout},
	}
}

// claudeRequest représente la structure de la requête à l'API Claude
type claudeRequest struct {
	Model     string    `json:"model"`
	Messages  []message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse représente la structure de la réponse de l'API Claude
type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// Translate implémente la méthode de l'interface LLMClient pour Claude
func (c *ClaudeClient) Translate(ctx context.Context, content, sourceLang, targetLang, additionalInstructions string) (string, error) {
	prompt := fmt.Sprintf(`Translate the following text from %s to %s. %s

Text to translate:
%s

Translated text:`, sourceLang, targetLang, additionalInstructions, content)

	reqBody := claudeRequest{
		Model: c.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: c.ContextSize,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ClaudeAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to Claude API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API returned non-OK status: %d", resp.StatusCode)
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", fmt.Errorf("error decoding Claude API response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in Claude API response")
	}

	return claudeResp.Content[0].Text, nil
}
