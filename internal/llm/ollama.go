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

type OllamaClient struct {
	host        string
	port        string
	model       string
	contextSize int
	timeout     time.Duration
	httpClient  *http.Client
}

func NewOllamaClient(host, port, model string, contextSize int, timeout time.Duration) *OllamaClient {
	return &OllamaClient{
		host:        host,
		port:        port,
		model:       model,
		contextSize: contextSize,
		timeout:     timeout,
		httpClient:  &http.Client{Timeout: timeout},
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

	logger.Debug(fmt.Sprintf("Prepared prompt for Ollama API:\n%s", prompt))

	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}
	reqBody.Options.NumCtx = c.contextSize

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	url := fmt.Sprintf("http://%s:%s/api/generate", c.host, c.port)

	var resp *http.Response
	var responseBody []byte
	maxRetries := 5
	baseDelay := 20 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", nil, fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		logger.Info(fmt.Sprintf("Sending request to Ollama API (Attempt %d of %d)", attempt+1, maxRetries))
		resp, err = c.httpClient.Do(req)
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
		return "", nil, fmt.Errorf("failed to get response from Ollama API after %d attempts", maxRetries)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(responseBody, &ollamaResp); err != nil {
		return "", nil, fmt.Errorf("error decoding Ollama API response: %w", err)
	}

	// Mettre à jour le contexte d'analyse
	updatedContext, err := UpdateAnalysisContext(ollamaResp.Response, analysisContext)
	if err != nil {
		logger.Error(fmt.Sprintf("Error updating analysis context: %v", err))
		return "", nil, fmt.Errorf("error updating analysis context: %w", err)
	}

	logger.Debug(fmt.Sprintf("Ollama API response:\n%s", ollamaResp.Response))
	logger.Info(fmt.Sprintf("Analysis completed successfully (response length: %d characters)", len(ollamaResp.Response)))
	return ollamaResp.Response, updatedContext, nil
}
