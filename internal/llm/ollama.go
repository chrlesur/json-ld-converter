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
func (c *OllamaClient) Analyze(ctx context.Context, content string) (string, error) {
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
