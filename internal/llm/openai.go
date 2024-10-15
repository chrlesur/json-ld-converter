package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/chrlesur/json-ld-converter/internal/logger"
	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client      *openai.Client
	model       string
	contextSize int
	timeout     time.Duration
}

func NewOpenAIClient(apiKey, model string, contextSize int, timeout time.Duration) *OpenAIClient {
	return &OpenAIClient{
		client:      openai.NewClient(apiKey),
		model:       model,
		contextSize: contextSize,
		timeout:     timeout,
	}
}

func (c *OpenAIClient) Analyze(ctx context.Context, content string, analysisContext *AnalysisContext) (string, *AnalysisContext, error) {
	logger.Debug("Starting analysis with OpenAI API")
	prompt := BuildPromptWithContext(content, analysisContext)

	logger.Debug(fmt.Sprintf("Prepared prompt for OpenAI API:\n%s", prompt))

	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens: c.contextSize,
	}

	var resp openai.ChatCompletionResponse
	var err error

	maxRetries := 5
	baseDelay := 20 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		logger.Info(fmt.Sprintf("Sending request to OpenAI API (Attempt %d of %d)", attempt+1, maxRetries))
		resp, err = c.client.CreateChatCompletion(ctxWithTimeout, req)
		if err == nil {
			break
		}

		logger.Warning(fmt.Sprintf("Attempt %d failed: %v", attempt+1, err))
		if attempt < maxRetries-1 {
			delay := baseDelay + time.Duration(attempt)*20*time.Second
			logger.Info(fmt.Sprintf("Retrying in %v", delay))
			time.Sleep(delay)
		}
	}

	if err != nil {
		logger.Error(fmt.Sprintf("All attempts failed. Last error: %v", err))
		return "", nil, fmt.Errorf("failed to get response from OpenAI API after %d attempts: %w", maxRetries, err)
	}

	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("no content in OpenAI API response")
	}

	responseText := resp.Choices[0].Message.Content

	// Mettre à jour le contexte d'analyse
	updatedContext, err := UpdateAnalysisContext(responseText, analysisContext)
	if err != nil {
		logger.Error(fmt.Sprintf("Error updating analysis context: %v", err))
		return "", nil, fmt.Errorf("error updating analysis context: %w", err)
	}

	logger.Debug(fmt.Sprintf("OpenAI API response:\n%s", responseText))
	logger.Info(fmt.Sprintf("Analysis completed successfully (response length: %d characters)", len(responseText)))
	return responseText, updatedContext, nil
}
