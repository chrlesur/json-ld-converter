package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	prompt := `Analysez le contenu fourni (représentant une partie d'un document plus large) et identifiez les principaux triplets entité-relation-attribut présents dans le texte. Concentrez-vous sur les concepts et relations importants au niveau du paragraphe, en gardant la chronologie des événements.

Instructions :
1. Analysez chaque paragraphe du chunk en détail.
2. Identifiez les triplets les plus pertinents et significatifs, en vous concentrant sur les idées principales et les informations clés.
3. Pour chaque triplet qui représente un fait à un moment donné, indiquez un lien vers l'événement précédent et suivant s'ils existent dans le même chunk.
4. Présentez les résultats sous forme de liste de triplets, un par ligne, séparés par des tabulations.

Format de réponse attendu :
"Entité principale"	"Relation importante"	"Attribut ou entité liée significative"	"Événement précédent (si applicable)"	"Événement suivant (si applicable)"
...

Assurez-vous que :
- Chaque triplet représente une information importante extraite du texte fourni.
- Les concepts, relations et attributs identifiés sont pertinents pour la compréhension globale du document.
- Les liens vers les événements précédents et suivants sont inclus uniquement pour les faits à un moment donné.
- Votre analyse capture l'essence du contenu et la séquence des informations telles qu'elles apparaissent dans le document.

IMPORTANT : Ne renvoyez que la liste des triplets avec leurs informations de séquence, sans aucun texte explicatif ou commentaire supplémentaire. L'application s'attend à recevoir uniquement les triplets bruts pour pouvoir les traiter correctement.

Contenu à analyser :
` + content

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

	logger.Info("Sending request to Claude API")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Error sending request to Claude API: %v", err))
		return "", fmt.Errorf("error sending request to Claude API: %w", err)
	}
	defer resp.Body.Close()
	logger.Debug(fmt.Sprintf("Received response from Claude API (status: %d)", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Claude API returned non-OK status: %d", resp.StatusCode))
		return "", fmt.Errorf("Claude API returned non-OK status: %d", resp.StatusCode)
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
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
