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

type OllamaClient struct {
	Host        string
	Port        string
	Model       string
	ContextSize int
	Timeout     time.Duration
	HTTPClient  *http.Client
}

func NewOllamaClient(host, port, model string, contextSize int, timeout time.Duration) *OllamaClient {
	logger.Info(fmt.Sprintf("Creating new OllamaClient with model: %s, contextSize: %d, timeout: %v", model, contextSize, timeout))
	return &OllamaClient{
		Host:        host,
		Port:        port,
		Model:       model,
		ContextSize: contextSize,
		Timeout:     timeout,
		HTTPClient:  &http.Client{Timeout: timeout},
	}
}

type ollamaRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Stream  bool   `json:"stream"`
	Options struct {
		NumCtx int `json:"num_ctx"`
	} `json:"options"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (c *OllamaClient) Analyze(ctx context.Context, content string, analysisContext *AnalysisContext) (string, *AnalysisContext, error) {
	logger.Debug("Starting analysis with Ollama API")
	prompt := BuildPromptWithContext(content, analysisContext)

	logger.Debug(fmt.Sprintf("Prepared prompt for Ollama API (%d tokens)", tokenizer.CountTokens(prompt)))

	reqBody := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}
	reqBody.Options.NumCtx = c.ContextSize

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error(fmt.Sprintf("Error marshaling request body: %v", err))
		return "", nil, fmt.Errorf("error marshaling request body: %w", err)
	}
	logger.Debug(fmt.Sprintf("Request body marshaled successfully (size: %d bytes)", len(jsonData)))

	url := fmt.Sprintf("http://%s:%s/api/generate", c.Host, c.Port)

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

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating request: %v", err))
			return "", nil, fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		logger.Info(fmt.Sprintf("Sending request to Ollama API (Attempt %d of %d, Timeout: %v)", attempt+1, maxRetries, currentTimeout))
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
			logger.Warning(fmt.Sprintf("Attempt %d failed with status code %d: %s", attempt+1, resp.StatusCode, string(responseBody)))
		}
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to get successful response from Ollama API after %d attempts", maxRetries)
	}

	logger.Debug(fmt.Sprintf("Received response from Ollama API (status: %d)", resp.StatusCode))

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(responseBody, &ollamaResp); err != nil {
		logger.Error(fmt.Sprintf("Error decoding Ollama API response: %v", err))
		return "", nil, fmt.Errorf("error decoding Ollama API response: %w", err)
	}
	logger.Debug("Ollama API response decoded successfully")

	responseText := ollamaResp.Response

	logger.Info(fmt.Sprintf("API Response received (%d tokens)", tokenizer.CountTokens(responseText)))
	logger.Debug(fmt.Sprintf("API Response content : %s", responseText))

	// Mettre à jour le contexte d'analyse
	updatedContext, err := UpdateAnalysisContext(responseText, analysisContext)
	if err != nil {
		logger.Error(fmt.Sprintf("Error updating analysis context: %v", err))
		return "", nil, fmt.Errorf("error updating analysis context: %w", err)
	}

	logger.Debug(fmt.Sprintf("Ollama API response:\n%s", responseText))
	logger.Debug(fmt.Sprintf("Analysis completed successfully (response length: %d characters)", len(responseText)))
	return responseText, updatedContext, nil
}
