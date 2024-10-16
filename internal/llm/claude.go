package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/chrlesur/json-ld-converter/internal/logger"
	"github.com/chrlesur/json-ld-converter/pkg/tokenizer"
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
func (c *ClaudeClient) Analyze(ctx context.Context, content string, analysisContext *AnalysisContext) (string, *AnalysisContext, error) {
	logger.Debug("Starting analysis with Claude API")
	prompt := BuildPromptWithContext(content, analysisContext)

	logger.Info(fmt.Sprintf("Prepared prompt for Claude API (%d tokens)", tokenizer.CountTokens(prompt)))

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
		return "", nil, fmt.Errorf("error marshaling request body: %w", err)
	}
	logger.Debug(fmt.Sprintf("Request body marshaled successfully (size: %d bytes)", len(jsonData)))

	var resp *http.Response
	var responseBody []byte
	maxRetries := 5
	baseTimeout := c.Timeout
	backoffFactor := 2 // Facteur pour le backoff exponentiel

	for attempt := 0; attempt < maxRetries; attempt++ {
		currentTimeout := baseTimeout * time.Duration(math.Pow(float64(backoffFactor), float64(attempt)))

		// Créer un nouveau client HTTP avec le timeout actuel
		clientWithTimeout := &http.Client{
			Timeout: currentTimeout,
		}

		// Utiliser le contexte parent sans timeout supplémentaire
		req, err := http.NewRequestWithContext(ctx, "POST", ClaudeAPIURL, bytes.NewBuffer(jsonData))
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating request: %v", err))
			return "", nil, fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", c.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		logger.Info(fmt.Sprintf("Sending request to Claude API (Attempt %d of %d, Timeout: %v)", attempt+1, maxRetries, currentTimeout))
		resp, err = clientWithTimeout.Do(req)

		if err == nil {
			defer resp.Body.Close()
			responseBody, err = ioutil.ReadAll(resp.Body)
			if err == nil && resp.StatusCode == http.StatusOK {
				break
			}
		}

		if err != nil {
			logger.Warning(fmt.Sprintf("Attempt %d failed: %v", attempt+1, err))
			if attempt < maxRetries-1 {
				retryDelay := time.Duration(math.Pow(float64(backoffFactor), float64(attempt))) * time.Second
				logger.Info(fmt.Sprintf("Retrying in %v", retryDelay))
				time.Sleep(retryDelay)
			}
		} else {
			// Si nous avons une réponse mais pas un statut 200, loguer le corps de la réponse
			logger.Warning(fmt.Sprintf("Attempt %d failed with status code %d: %s", attempt+1, resp.StatusCode, string(responseBody)))
		}
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to get successful response from Claude API after %d attempts", maxRetries)
	}

	logger.Debug(fmt.Sprintf("Received response from Claude API (status: %d)", resp.StatusCode))

	var claudeResp claudeResponse
	if err := json.Unmarshal(responseBody, &claudeResp); err != nil {
		logger.Error(fmt.Sprintf("Error decoding Claude API response: %v", err))
		return "", nil, fmt.Errorf("error decoding Claude API response: %w", err)
	}
	logger.Debug("Claude API response decoded successfully")

	if len(claudeResp.Content) == 0 {
		logger.Error("No content in Claude API response")
		return "", nil, fmt.Errorf("no content in Claude API response")
	}

	responseText := claudeResp.Content[0].Text

	logger.Info(fmt.Sprintf("API Response : %s (%d tokens)", responseText, tokenizer.CountTokens(responseText)))

	// Mettre à jour le contexte d'analyse
	updatedContext, err := UpdateAnalysisContext(responseText, analysisContext)
	if err != nil {
		logger.Error(fmt.Sprintf("Error updating analysis context: %v", err))
		return "", nil, fmt.Errorf("error updating analysis context: %w", err)
	}

	logger.Debug(fmt.Sprintf("Claude API response:\n%s", responseText))
	logger.Info(fmt.Sprintf("Analysis completed successfully (response length: %d characters)", len(responseText)))
	return responseText, updatedContext, nil
}
