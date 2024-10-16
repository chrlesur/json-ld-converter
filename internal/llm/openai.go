package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/chrlesur/json-ld-converter/internal/logger"
	"github.com/chrlesur/json-ld-converter/pkg/tokenizer"
	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	APIKey      string
	Model       string
	ContextSize int
	Timeout     time.Duration
}

func NewOpenAIClient(apiKey, model string, contextSize int, timeout time.Duration) *OpenAIClient {
	logger.Info(fmt.Sprintf("Creating new OpenAIClient with model: %s, contextSize: %d, timeout: %v", model, contextSize, timeout))
	return &OpenAIClient{
		APIKey:      apiKey,
		Model:       model,
		ContextSize: contextSize,
		Timeout:     timeout,
	}
}

func (c *OpenAIClient) Analyze(ctx context.Context, content string, analysisContext *AnalysisContext) (string, *AnalysisContext, error) {
	logger.Debug("Starting analysis with OpenAI API")
	prompt := BuildPromptWithContext(content, analysisContext)

	logger.Debug(fmt.Sprintf("Prepared prompt for OpenAI API (%d tokens)", tokenizer.CountTokens(prompt)))

	req := openai.ChatCompletionRequest{
		Model: c.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens: c.ContextSize,
	}
	var resp openai.ChatCompletionResponse
	var err error

	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Error marshaling request body: %v", err))
		return "", nil, fmt.Errorf("error marshaling request body: %w", err)
	}
	logger.Debug(fmt.Sprintf("Request body marshaled successfully (size: %d bytes)", len(jsonData)))

	maxRetries := 5
	baseTimeout := c.Timeout
	backoffFactor := 2 // Facteur pour le backoff exponentiel

	for attempt := 0; attempt < maxRetries; attempt++ {
		currentTimeout := baseTimeout * time.Duration(math.Pow(float64(backoffFactor), float64(attempt)))

		// Créer un nouveau client OpenAI avec le timeout actuel
		config := openai.DefaultConfig(c.APIKey)
		config.HTTPClient = &http.Client{Timeout: currentTimeout}
		client := openai.NewClientWithConfig(config)

		logger.Info(fmt.Sprintf("Sending request to OpenAI API (Attempt %d of %d, Timeout: %v)", attempt+1, maxRetries, currentTimeout))

		ctxWithTimeout, cancel := context.WithTimeout(ctx, currentTimeout)
		resp, err = client.CreateChatCompletion(ctxWithTimeout, req)
		cancel()

		if err == nil {
			break
		}

		logger.Warning(fmt.Sprintf("Attempt %d failed: %v", attempt+1, err))
		if attempt < maxRetries-1 {
			retryDelay := time.Duration(math.Pow(float64(backoffFactor), float64(attempt))) * time.Second
			logger.Info(fmt.Sprintf("Retrying in %v", retryDelay))
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		logger.Error(fmt.Sprintf("All attempts failed. Last error: %v", err))
		return "", nil, fmt.Errorf("failed to get response from OpenAI API after %d attempts: %w", maxRetries, err)
	}

	if len(resp.Choices) == 0 {
		logger.Error("No content in OpenAI API response")
		return "", nil, fmt.Errorf("no content in OpenAI API response")
	}

	responseText := resp.Choices[0].Message.Content

	logger.Info(fmt.Sprintf("API Response received (%d tokens)", resp.Usage.CompletionTokens))
	logger.Debug(fmt.Sprintf("API Response content : %s", responseText))

	// Nettoyez la réponse
	cleanedResponse := cleanJSONResponse(responseText)

	// Parse le JSON nettoyé
	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(cleanedResponse), &jsonResponse)
	if err != nil {
		return "", nil, fmt.Errorf("erreur lors du parsing de la réponse JSON : %w", err)
	}

	// Mettre à jour le contexte d'analyse
	updatedContext, err := UpdateAnalysisContext(responseText, analysisContext)
	if err != nil {
		logger.Error(fmt.Sprintf("Error updating analysis context: %v", err))
		return "", nil, fmt.Errorf("error updating analysis context: %w", err)
	}

	logger.Debug(fmt.Sprintf("OpenAI API response:\n%s", responseText))
	logger.Debug(fmt.Sprintf("Analysis completed successfully (response length: %d characters)", len(responseText)))
	return responseText, updatedContext, nil
}

func cleanJSONResponse(response string) string {
	// Supprimez les backticks et le mot "json" s'ils sont présents
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimSuffix(response, "```")

	// Supprimez les espaces blancs au début et à la fin
	response = strings.TrimSpace(response)

	return response
}
