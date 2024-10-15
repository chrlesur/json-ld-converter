package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chrlesur/json-ld-converter/internal/logger"
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
	logger.Info(fmt.Sprintf("Creating new ClaudeClient with model: %s, contextSize: %d, timeout: %v", model, contextSize, timeout))
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

// Analyze implémente la méthode de l'interface LLMClient pour Claude
func (c *ClaudeClient) Analyze(ctx context.Context, content string) (string, error) {
	logger.Debug("Starting analysis with Claude API")
	prompt := `` + content

	logger.Debug(fmt.Sprintf("Prepared prompt for Claude API:\n%s", prompt))

	reqBody := claudeRequest{
		Model: c.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: c.ContextSize,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error(fmt.Sprintf("Error marshaling request body: %v", err))
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}
	logger.Debug(fmt.Sprintf("Request body marshaled successfully (size: %d bytes)", len(jsonData)))

	var resp *http.Response
	var responseBody []byte
	maxRetries := 5
	baseDelay := 20 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", ClaudeAPIURL, bytes.NewBuffer(jsonData))
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating request: %v", err))
			return "", fmt.Errorf("error creating request: %w", err)
		}
		logger.Debug("HTTP request created successfully")

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", c.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		logger.Debug("Request headers set")

		logger.Info(fmt.Sprintf("Sending request to Claude API (Attempt %d of %d)", attempt+1, maxRetries))
		resp, err = c.HTTPClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			responseBody, err = ioutil.ReadAll(resp.Body)
			if err == nil && resp.StatusCode == http.StatusOK {
				break
			}
		}

		logger.Warning(fmt.Sprintf("Attempt %d failed: %v", attempt+1, err))
		if attempt < maxRetries-1 {
			delay := baseDelay + time.Duration(attempt)*20*time.Second
			logger.Info(fmt.Sprintf("Retrying in %v", delay))
			time.Sleep(delay)
		}
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("All attempts failed. Last error: %v", err))
		return "", fmt.Errorf("failed to get response from Claude API after %d attempts", maxRetries)
	}

	logger.Debug(fmt.Sprintf("Received response from Claude API (status: %d)", resp.StatusCode))

	var claudeResp claudeResponse
	if err := json.Unmarshal(responseBody, &claudeResp); err != nil {
		logger.Error(fmt.Sprintf("Error decoding Claude API response: %v", err))
		return "", fmt.Errorf("error decoding Claude API response: %w", err)
	}
	logger.Debug("Claude API response decoded successfully")

	if len(claudeResp.Content) == 0 {
		logger.Error("No content in Claude API response")
		return "", fmt.Errorf("no content in Claude API response")
	}

	responseText := claudeResp.Content[0].Text
	logger.Debug(fmt.Sprintf("Claude API response:\n%s", responseText))
	logger.Info(fmt.Sprintf("Analysis completed successfully (response length: %d characters)", len(responseText)))
	return responseText, nil
}
