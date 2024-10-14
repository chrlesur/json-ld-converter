package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
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
func (c *OpenAIClient) Analyze(ctx context.Context, content string) (string, error) {
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
