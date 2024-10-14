package llm

import (
    "context"
    "fmt"
    "github.com/sashabaranov/go-openai"
    "time"
)

// OpenAIClient implémente l'interface LLMClient pour les modèles GPT d'OpenAI
type OpenAIClient struct {
    client      *openai.Client
    model       string
    contextSize int
    timeout     time.Duration
}

// NewOpenAIClient crée et retourne une nouvelle instance de OpenAIClient
func NewOpenAIClient(apiKey, model string, contextSize int, timeout time.Duration) *OpenAIClient {
    return &OpenAIClient{
        client:      openai.NewClient(apiKey),
        model:       model,
        contextSize: contextSize,
        timeout:     timeout,
    }
}

// Translate implémente la méthode de l'interface LLMClient pour OpenAI
func (c *OpenAIClient) Translate(ctx context.Context, content, sourceLang, targetLang, additionalInstructions string) (string, error) {
    prompt := fmt.Sprintf(`Translate the following text from %s to %s. %s

Text to translate:
%s

Translated text:`, sourceLang, targetLang, additionalInstructions, content)

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

    ctx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()

    resp, err := c.client.CreateChatCompletion(ctx, req)
    if err != nil {
        return "", fmt.Errorf("error calling OpenAI API: %w", err)
    }

    if len(resp.Choices) == 0 {
        return "", fmt.Errorf("no content in OpenAI API response")
    }

    return resp.Choices[0].Message.Content, nil
}