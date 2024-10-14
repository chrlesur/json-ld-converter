package llm

import (
	"fmt"
	"os"
	"time"

	"github.com/chrlesur/json-ld-converter/internal/config"
)

// NewLLMClient crée et retourne le client LLM approprié en fonction du type de moteur spécifié
func NewLLMClient(cfg *config.Config) (LLMClient, error) {
    switch cfg.Conversion.Engine {
    case "claude":
        apiKey := os.Getenv("CLAUDE_API_KEY")
        if apiKey == "" {
            return nil, fmt.Errorf("CLAUDE_API_KEY environment variable is not set")
        }
        return NewClaudeClient(apiKey, cfg.Conversion.Model, cfg.Conversion.ContextSize, time.Duration(cfg.Conversion.Timeout)*time.Second), nil
    case "openai":
        apiKey := os.Getenv("OPENAI_API_KEY")
        if apiKey == "" {
            return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
        }
        return NewOpenAIClient(apiKey, cfg.Conversion.Model, cfg.Conversion.ContextSize, time.Duration(cfg.Conversion.Timeout)*time.Second), nil
    case "ollama":
        return NewOllamaClient(cfg.Conversion.OllamaHost, cfg.Conversion.OllamaPort, cfg.Conversion.Model, cfg.Conversion.ContextSize, time.Duration(cfg.Conversion.Timeout)*time.Second), nil
    case "aiyou":
        client := NewAIYOUClient(cfg.Conversion.AIYOUAssistantID, time.Duration(cfg.Conversion.Timeout)*time.Second)
        email := os.Getenv("AIYOU_EMAIL")
        password := os.Getenv("AIYOU_PASSWORD")
        if email == "" || password == "" {
            return nil, fmt.Errorf("AIYOU_EMAIL or AIYOU_PASSWORD environment variable is not set")
        }
        err := client.Login(email, password)
        if err != nil {
            return nil, fmt.Errorf("failed to login to AI.YOU: %w", err)
        }
        return client, nil
    default:
        return nil, fmt.Errorf("unsupported LLM engine: %s", cfg.Conversion.Engine)
    }
}