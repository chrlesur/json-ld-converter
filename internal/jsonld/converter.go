package jsonld

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chrlesur/json-ld-converter/internal/llm"
	"github.com/chrlesur/json-ld-converter/internal/parser"
	"github.com/chrlesur/json-ld-converter/internal/schema"
	"github.com/chrlesur/json-ld-converter/pkg/tokenizer"
)

type Converter struct {
	schemaVocabulary       *schema.Vocabulary
	llmClient              llm.TranslationClient
	maxTokens              int
	additionalInstructions string
}

func NewConverter(vocabulary *schema.Vocabulary, client llm.TranslationClient, maxTokens int, instructions string) *Converter {
	return &Converter{
		schemaVocabulary:       vocabulary,
		llmClient:              client,
		maxTokens:              maxTokens,
		additionalInstructions: instructions,
	}
}

func (c *Converter) Convert(segment *parser.DocumentSegment) (map[string]interface{}, error) {
    // Vérifier si le contenu dépasse la limite de tokens
    if tokenizer.CountTokens(segment.Content) > c.maxTokens {
        return nil, &TokenLimitError{Limit: c.maxTokens, Count: tokenizer.CountTokens(segment.Content)}
    }

    // Initialiser la structure JSON-LD de base
    jsonLD := map[string]interface{}{
        "@context": "https://schema.org",
    }

    // Utiliser le LLM pour enrichir la conversion
    enrichedContent, err := c.enrichContentWithLLM(segment.Content)
    if err != nil {
        return nil, &ConversionError{Stage: "enrichissement", Err: err}
    }

    // Déterminer le type principal
    mainType, err := c.determineMainType(enrichedContent)
    if err != nil {
        // Stratégie de repli : utiliser "Thing" comme type par défaut
        mainType = "Thing"
        jsonLD["@type"] = mainType
    }

    // Gérer les structures imbriquées
    nestedContent, err := c.handleNestedStructures(enrichedContent, mainType)
    if err != nil {
        // Stratégie de repli : utiliser une structure plate si la structure imbriquée échoue
        nestedContent, err = c.extractProperties(enrichedContent, mainType)
        if err != nil {
            return nil, &ConversionError{Stage: "extraction des propriétés", Err: err}
        }
    }

    // Ajouter le contenu imbriqué à la structure JSON-LD
    for key, value := range nestedContent {
        jsonLD[key] = value
    }

    // Appliquer les instructions supplémentaires
    jsonLD, err = c.applyAdditionalInstructions(jsonLD)
    if err != nil {
        // Ignorer l'erreur des instructions supplémentaires et continuer avec le JSON-LD non modifié
        fmt.Printf("Avertissement : impossible d'appliquer les instructions supplémentaires : %v\n", err)
    }

    // Vérifier la limite de tokens
    if err := c.checkTokenLimit(jsonLD); err != nil {
        return nil, &TokenLimitError{Limit: c.maxTokens, Count: tokenizer.CountTokens(fmt.Sprintf("%v", jsonLD))}
    }

    return jsonLD, nil
}

func (c *Converter) enrichContentWithLLM(content string) (string, error) {
	prompt := fmt.Sprintf("Analysez et enrichissez sémantiquement le contenu suivant : %s", content)
	enrichedContent, err := c.llmClient.Translate(prompt, "français", "français", "")
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'appel au LLM : %w", err)
	}
	return enrichedContent, nil
}

func (c *Converter) mapToSchemaOrg(content string) (map[string]interface{}, error) {
	// Analyser le contenu pour déterminer le type principal
	mainType, err := c.determineMainType(content)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la détermination du type principal : %w", err)
	}

	// Initialiser la structure JSON-LD avec le type principal
	jsonLD := map[string]interface{}{
		"@type": mainType,
	}

	// Extraire et mapper les propriétés pertinentes
	properties, err := c.extractProperties(content, mainType)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'extraction des propriétés : %w", err)
	}

	// Ajouter les propriétés à la structure JSON-LD
	for key, value := range properties {
		jsonLD[key] = value
	}

	return jsonLD, nil
}

func (c *Converter) determineMainType(content string) (string, error) {
	// Utiliser le LLM pour déterminer le type principal du contenu
	prompt := fmt.Sprintf("Déterminez le type Schema.org le plus approprié pour le contenu suivant : %s", content)
	response, err := c.llmClient.Translate(prompt, "français", "français", "")
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'appel au LLM : %w", err)
	}

	// Vérifier si le type retourné existe dans le vocabulaire Schema.org
	if !c.schemaVocabulary.TypeExists(response) {
		return "Thing", nil // Utiliser "Thing" comme type par défaut si le type n'est pas reconnu
	}

	return response, nil
}

func (c *Converter) extractProperties(content string, mainType string) (map[string]interface{}, error) {
	properties := make(map[string]interface{})

	// Obtenir les propriétés applicables pour le type principal
	applicableProperties := c.schemaVocabulary.GetPropertiesForType(mainType)

	// Utiliser le LLM pour extraire les valeurs des propriétés applicables
	for _, prop := range applicableProperties {
		prompt := fmt.Sprintf("Extrayez la valeur de la propriété '%s' pour un objet de type '%s' à partir du contenu suivant : %s", prop, mainType, content)
		value, err := c.llmClient.Translate(prompt, "français", "français", "")
		if err != nil {
			return nil, fmt.Errorf("erreur lors de l'extraction de la propriété '%s' : %w", prop, err)
		}

		if value != "" {
			properties[prop] = value
		}
	}

	return properties, nil
}

func (c *Converter) checkTokenLimit(jsonLD map[string]interface{}) error {
	jsonString, err := json.Marshal(jsonLD)
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation JSON : %w", err)
	}

	tokenCount := tokenizer.CountTokens(string(jsonString))
	if tokenCount > c.maxTokens {
		return fmt.Errorf("la limite de tokens (%d) a été dépassée : %d tokens", c.maxTokens, tokenCount)
	}

	return nil
}

func (c *Converter) handleNestedStructures(content string, mainType string) (map[string]interface{}, error) {
    jsonLD := make(map[string]interface{})
    jsonLD["@type"] = mainType

    properties, err := c.extractProperties(content, mainType)
    if err != nil {
        return nil, &ConversionError{Stage: "extraction des propriétés", Err: err}
    }

    for key, value := range properties {
        if c.schemaVocabulary.IsObjectProperty(mainType, key) {
            nestedType, err := c.schemaVocabulary.GetExpectedType(mainType, key)
            if err != nil {
                // Stratégie de repli : utiliser "Thing" comme type par défaut pour les propriétés d'objet
                nestedType = "Thing"
            }
            
            nestedContent, err := c.extractNestedContent(content, key)
            if err != nil {
                // Stratégie de repli : utiliser la valeur extraite comme contenu texte simple
                jsonLD[key] = value
                continue
            }

            nestedStructure, err := c.handleNestedStructures(nestedContent, nestedType)
            if err != nil {
                // Stratégie de repli : utiliser une structure plate pour le contenu imbriqué
                jsonLD[key] = map[string]interface{}{
                    "@type": nestedType,
                    "text":  nestedContent,
                }
            } else {
                jsonLD[key] = nestedStructure
            }
        } else {
            jsonLD[key] = value
        }
    }

    return jsonLD, nil
}

func (c *Converter) extractNestedContent(content string, property string) (string, error) {
	prompt := fmt.Sprintf("Extrayez le contenu spécifique à la propriété '%s' à partir du texte suivant : %s", property, content)
	nestedContent, err := c.llmClient.Translate(prompt, "français", "français", "")
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'extraction du contenu imbriqué : %w", err)
	}
	return nestedContent, nil
}

func (c *Converter) applyAdditionalInstructions(jsonLD map[string]interface{}) (map[string]interface{}, error) {
	if c.additionalInstructions == "" {
		return jsonLD, nil
	}

	prompt := fmt.Sprintf("Appliquez les instructions suivantes au JSON-LD : %s\nJSON-LD actuel : %v",
		c.additionalInstructions, jsonLD)

	response, err := c.llmClient.Translate(prompt, "français", "français", "")
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'application des instructions supplémentaires : %w", err)
	}

	var modifiedJsonLD map[string]interface{}
	err = json.Unmarshal([]byte(response), &modifiedJsonLD)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation du JSON-LD modifié : %w", err)
	}

	return modifiedJsonLD, nil
}

func (c *Converter) splitContent(content string) []string {
	tokens := tokenizer.Tokenize(content)
	var segments []string
	var currentSegment []string

	for _, token := range tokens {
		if len(currentSegment)+1 > c.maxTokens/2 { // Utilisation de la moitié de la limite pour laisser de la place à la structure JSON-LD
			segments = append(segments, strings.Join(currentSegment, " "))
			currentSegment = []string{}
		}
		currentSegment = append(currentSegment, token)
	}

	if len(currentSegment) > 0 {
		segments = append(segments, strings.Join(currentSegment, " "))
	}

	return segments
}

func (c *Converter) convertLargeDocument(content string) ([]map[string]interface{}, error) {
	segments := c.splitContent(content)
	var results []map[string]interface{}

	for i, segment := range segments {
		jsonLD, err := c.convertSegment(segment)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la conversion du segment %d : %w", i, err)
		}

		// Ajouter des métadonnées pour indiquer la segmentation
		jsonLD["segment"] = map[string]interface{}{
			"index": i + 1,
			"total": len(segments),
		}

		results = append(results, jsonLD)
	}

	return results, nil
}

func (c *Converter) convertSegment(segment string) (map[string]interface{}, error) {
	// Utiliser la méthode Convert existante
	return c.Convert(&parser.DocumentSegment{Content: segment})
}
