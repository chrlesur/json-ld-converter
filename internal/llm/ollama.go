package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaClient implémente l'interface LLMClient pour les modèles Ollama
type OllamaClient struct {
	host        string
	port        string
	model       string
	contextSize int
	timeout     time.Duration
	httpClient  *http.Client
}

// NewOllamaClient crée et retourne une nouvelle instance de OllamaClient
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

// Translate implémente la méthode de l'interface LLMClient pour Ollama
func (c *OllamaClient) Translate(ctx context.Context, content, sourceLang, targetLang, additionalInstructions string) (string, error) {
	prompt := fmt.Sprintf(`Translate the following text from %s to %s. %s

Text to translate:
%s

Translated text:`, sourceLang, targetLang, additionalInstructions, content)

	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}
	reqBody.Options.NumCtx = c.contextSize

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	url := fmt.Sprintf("http://%s:%s/api/generate", c.host, c.port)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API returned non-OK status: %d", resp.StatusCode)
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("error decoding Ollama API response: %w", err)
	}

	return ollamaResp.Response, nil
}
