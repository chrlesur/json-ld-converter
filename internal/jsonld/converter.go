package jsonld

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/chrlesur/json-ld-converter/internal/llm"
	"github.com/chrlesur/json-ld-converter/internal/logger"
	"github.com/chrlesur/json-ld-converter/internal/parser"
	"github.com/chrlesur/json-ld-converter/internal/schema"
	"github.com/chrlesur/json-ld-converter/pkg/tokenizer"
)

type Converter struct {
	schemaOrg              *schema.SchemaOrg
	llmClient              llm.LLMClient
	maxTokens              int
	additionalInstructions string
	analysisContext        *llm.AnalysisContext
}

func NewConverter(schemaOrg *schema.SchemaOrg, client llm.LLMClient, maxTokens int, instructions string) *Converter {
	logger.Info("Creating new Converter instance")
	return &Converter{
		schemaOrg:              schemaOrg,
		llmClient:              client,
		maxTokens:              maxTokens,
		additionalInstructions: instructions,
		analysisContext:        &llm.AnalysisContext{},
	}
}

func (c *Converter) Convert(ctx context.Context, doc *parser.Document) (map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("Starting conversion process for document with %d tokens", tokenizer.CountTokens(doc.Content)))

	if tokenizer.CountTokens(doc.Content) > c.maxTokens {
		logger.Warning(fmt.Sprintf("Document exceeds token limit: %d tokens (limit: %d)", tokenizer.CountTokens(doc.Content), c.maxTokens))
		return nil, &TokenLimitError{Limit: c.maxTokens, Count: tokenizer.CountTokens(doc.Content)}
	}

	jsonLD := map[string]interface{}{
		"@context": "https://schema.org",
	}
	logger.Debug("Initialized base JSON-LD structure")

	logger.Info("Enriching content with LLM")
	enrichedContent, newContext, err := c.enrichContentWithLLM(ctx, doc.Content)
	if err != nil {
		logger.Error(fmt.Sprintf("Error enriching content with LLM: %v", err))
		return nil, &ConversionError{Stage: "enrichissement", Err: err}
	}
	c.analysisContext = newContext
	logger.Debug("Content successfully enriched by LLM")

	logger.Info("Determining main type")
	mainType, err := c.determineMainType(ctx, enrichedContent)
	if err != nil {
		logger.Warning(fmt.Sprintf("Error determining main type: %v. Falling back to 'Thing'", err))
		mainType = "Thing"
		jsonLD["@type"] = mainType
	}
	logger.Debug(fmt.Sprintf("Main type determined: %s", mainType))

	logger.Info("Handling nested structures")
	nestedContent, err := c.handleNestedStructures(ctx, enrichedContent, mainType)
	if err != nil {
		logger.Warning(fmt.Sprintf("Error handling nested structures: %v. Falling back to flat structure", err))
		nestedContent, err = c.extractProperties(ctx, enrichedContent, mainType)
		if err != nil {
			logger.Error(fmt.Sprintf("Error extracting properties: %v", err))
			return nil, &ConversionError{Stage: "extraction des propriétés", Err: err}
		}
	}
	logger.Debug("Nested structures handled successfully")

	for key, value := range nestedContent {
		jsonLD[key] = value
	}
	logger.Debug("Nested content added to JSON-LD structure")

	logger.Info("Applying additional instructions")
	jsonLD, err = c.applyAdditionalInstructions(ctx, jsonLD)
	if err != nil {
		logger.Warning(fmt.Sprintf("Unable to apply additional instructions: %v", err))
	} else {
		logger.Debug("Additional instructions applied successfully")
	}

	logger.Info("Checking final token limit")
	if err := c.checkTokenLimit(jsonLD); err != nil {
		logger.Error(fmt.Sprintf("Final JSON-LD exceeds token limit: %v", err))
		return nil, &TokenLimitError{Limit: c.maxTokens, Count: tokenizer.CountTokens(fmt.Sprintf("%v", jsonLD))}
	}
	logger.Debug("Final JSON-LD within token limit")

	logger.Info("Conversion process completed successfully")
	return jsonLD, nil
}

func (c *Converter) enrichContentWithLLM(ctx context.Context, content string) (string, *llm.AnalysisContext, error) {
	logger.Debug("Enriching content with LLM")
	prompt := llm.BuildPromptWithContext(content, c.analysisContext)
	logger.Debug(fmt.Sprintf("Prepared prompt for LLM:\n%s", prompt))

	enrichedContent, newContext, err := c.llmClient.Analyze(ctx, prompt, c.analysisContext)
	if err != nil {
		logger.Error(fmt.Sprintf("Error calling LLM for content enrichment: %v", err))
		return "", nil, fmt.Errorf("erreur lors de l'appel au LLM : %w", err)
	}
	logger.Debug("Content successfully enriched by LLM")
	return enrichedContent, newContext, nil
}

func (c *Converter) determineMainType(ctx context.Context, content string) (string, error) {
	logger.Debug("Determining main type")
	prompt := fmt.Sprintf("Déterminez le type Schema.org le plus approprié pour le contenu suivant. Répondez uniquement avec le nom du type, sans explication : %s", content)
	response, newContext, err := c.llmClient.Analyze(ctx, prompt, c.analysisContext)
	if err != nil {
		logger.Error(fmt.Sprintf("Error calling LLM for main type determination: %v", err))
		return "", fmt.Errorf("erreur lors de l'appel au LLM : %w", err)
	}
	c.analysisContext = newContext

	mainType := c.extractSchemaOrgType(response)

	if _, ok := c.schemaOrg.GetType(mainType); !ok {
		logger.Warning(fmt.Sprintf("Type '%s' not found in Schema.org vocabulary. Using 'Thing' as default", mainType))
		return "Thing", nil
	}

	logger.Debug(fmt.Sprintf("Main type determined: %s", mainType))
	return mainType, nil
}

func (c *Converter) extractSchemaOrgType(response string) string {
	logger.Debug(fmt.Sprintf("Extracting Schema.org type from LLM response: %s", response))

	// Liste des types Schema.org courants à rechercher
	commonTypes := []string{"Article", "Person", "Event", "Organization", "Place", "CreativeWork", "Thing"}

	// Vérifier si l'un des types courants est présent dans la réponse
	for _, t := range commonTypes {
		if strings.Contains(response, t) {
			logger.Debug(fmt.Sprintf("Found common Schema.org type: %s", t))
			return t
		}
	}

	// Si aucun type commun n'est trouvé, chercher le premier mot qui pourrait être un type
	words := strings.Fields(response)
	for _, word := range words {
		// Vérifier si le mot commence par une majuscule (potentiel type Schema.org)
		if len(word) > 0 && unicode.IsUpper(rune(word[0])) {
			logger.Debug(fmt.Sprintf("Extracted potential Schema.org type: %s", word))
			return word
		}
	}

	logger.Warning("No valid Schema.org type found in response. Defaulting to 'Thing'")
	return "Thing"
}

func (c *Converter) extractProperties(ctx context.Context, content string, mainType string) (map[string]interface{}, error) {
	logger.Debug(fmt.Sprintf("Extracting properties for type: %s", mainType))
	properties := make(map[string]interface{})

	// Obtenir les propriétés applicables pour le type principal
	schemaType, ok := c.schemaOrg.GetType(mainType)
	if !ok {
		logger.Error(fmt.Sprintf("Type not found in schema: %s", mainType))
		return nil, fmt.Errorf("type non trouvé dans le schéma : %s", mainType)
	}

	// Utiliser le LLM pour extraire les valeurs des propriétés applicables
	for _, prop := range schemaType.Properties {
		logger.Debug(fmt.Sprintf("Extracting property: %s", prop))
		prompt := fmt.Sprintf("Extrayez la valeur de la propriété '%s' pour un objet de type '%s' à partir du contenu suivant : %s", prop, mainType, content)
		value, _, err := c.llmClient.Analyze(ctx, prompt, &llm.AnalysisContext{})
		if err != nil {
			logger.Error(fmt.Sprintf("Error extracting property '%s': %v", prop, err))
			return nil, fmt.Errorf("erreur lors de l'extraction de la propriété '%s' : %w", prop, err)
		}

		if value != "" {
			properties[prop] = value
			logger.Debug(fmt.Sprintf("Property '%s' extracted with value: %s", prop, value))
		} else {
			logger.Debug(fmt.Sprintf("No value found for property: %s", prop))
		}
	}

	logger.Info(fmt.Sprintf("Extracted %d properties for type %s", len(properties), mainType))
	return properties, nil
}

func (c *Converter) handleNestedStructures(ctx context.Context, content string, mainType string) (map[string]interface{}, error) {
	logger.Debug(fmt.Sprintf("Handling nested structures for type: %s", mainType))
	jsonLD := make(map[string]interface{})
	jsonLD["@type"] = mainType

	properties, err := c.extractProperties(ctx, content, mainType)
	if err != nil {
		logger.Error(fmt.Sprintf("Error extracting properties: %v", err))
		return nil, &ConversionError{Stage: "extraction des propriétés", Err: err}
	}

	for key, value := range properties {
		logger.Debug(fmt.Sprintf("Processing property: %s", key))
		if c.isObjectProperty(mainType, key) {
			logger.Debug(fmt.Sprintf("Property %s is an object property", key))
			nestedType, err := c.getExpectedType(mainType, key)
			if err != nil {
				logger.Warning(fmt.Sprintf("Error getting expected type for property %s: %v. Using 'Thing' as default", key, err))
				// Stratégie de repli : utiliser "Thing" comme type par défaut pour les propriétés d'objet
				nestedType = "Thing"
			}

			nestedContent, err := c.extractNestedContent(ctx, content, key)
			if err != nil {
				logger.Warning(fmt.Sprintf("Error extracting nested content for property %s: %v. Using simple text value", key, err))
				// Stratégie de repli : utiliser la valeur extraite comme contenu texte simple
				jsonLD[key] = value
				continue
			}

			nestedStructure, err := c.handleNestedStructures(ctx, nestedContent, nestedType)
			if err != nil {
				logger.Warning(fmt.Sprintf("Error handling nested structure for property %s: %v. Using flat structure", key, err))
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

	logger.Info(fmt.Sprintf("Nested structures handled for type %s with %d properties", mainType, len(jsonLD)))
	return jsonLD, nil
}

func (c *Converter) isObjectProperty(typeName, propertyName string) bool {
	// Implémentez la logique pour déterminer si une propriété est un objet
	// Cela peut impliquer de vérifier le type de la propriété dans le schéma
	logger.Debug(fmt.Sprintf("Checking if property %s of type %s is an object property", propertyName, typeName))
	return false // Placeholder
}

func (c *Converter) getExpectedType(typeName, propertyName string) (string, error) {
	// Implémentez la logique pour obtenir le type attendu d'une propriété
	logger.Debug(fmt.Sprintf("Getting expected type for property %s of type %s", propertyName, typeName))
	return "", fmt.Errorf("not implemented")
}

func (c *Converter) extractNestedContent(ctx context.Context, content string, property string) (string, error) {
	logger.Debug(fmt.Sprintf("Extracting nested content for property: %s", property))
	prompt := fmt.Sprintf("Extrayez le contenu spécifique à la propriété '%s' à partir du texte suivant : %s", property, content)
	nestedContent, _, err := c.llmClient.Analyze(ctx, prompt, &llm.AnalysisContext{})
	if err != nil {
		logger.Error(fmt.Sprintf("Error extracting nested content for property %s: %v", property, err))
		return "", fmt.Errorf("erreur lors de l'extraction du contenu imbriqué : %w", err)
	}
	logger.Debug(fmt.Sprintf("Nested content extracted for property %s", property))
	return nestedContent, nil
}

func (c *Converter) applyAdditionalInstructions(ctx context.Context, jsonLD map[string]interface{}) (map[string]interface{}, error) {
	if c.additionalInstructions == "" {
		logger.Debug("No additional instructions to apply")
		return jsonLD, nil
	}

	logger.Info("Applying additional instructions")
	prompt := fmt.Sprintf("Appliquez les instructions suivantes au JSON-LD : %s\nJSON-LD actuel : %v",
		c.additionalInstructions, jsonLD)

	response, _, err := c.llmClient.Analyze(ctx, prompt, &llm.AnalysisContext{})
	if err != nil {
		logger.Error(fmt.Sprintf("Error applying additional instructions: %v", err))
		return nil, fmt.Errorf("erreur lors de l'application des instructions supplémentaires : %w", err)
	}

	var modifiedJsonLD map[string]interface{}
	err = json.Unmarshal([]byte(response), &modifiedJsonLD)
	if err != nil {
		logger.Error(fmt.Sprintf("Error unmarshaling modified JSON-LD: %v", err))
		return nil, fmt.Errorf("erreur lors de la désérialisation du JSON-LD modifié : %w", err)
	}

	logger.Debug("Additional instructions applied successfully")
	return modifiedJsonLD, nil
}

func (c *Converter) checkTokenLimit(jsonLD map[string]interface{}) error {
	logger.Debug("Checking token limit for final JSON-LD")
	jsonString, err := json.Marshal(jsonLD)
	if err != nil {
		logger.Error(fmt.Sprintf("Error marshaling JSON-LD: %v", err))
		return fmt.Errorf("erreur lors de la sérialisation JSON : %w", err)
	}

	tokenCount := tokenizer.CountTokens(string(jsonString))
	if tokenCount > c.maxTokens {
		logger.Warning(fmt.Sprintf("Token limit exceeded: %d tokens (limit: %d)", tokenCount, c.maxTokens))
		return fmt.Errorf("la limite de tokens (%d) a été dépassée : %d tokens", c.maxTokens, tokenCount)
	}

	logger.Debug(fmt.Sprintf("JSON-LD is within token limit: %d tokens", tokenCount))
	return nil
}

func (c *Converter) splitContent(content string) []string {
	logger.Debug(fmt.Sprintf("Splitting content into tokens (total length: %d)", len(content)))
	return tokenizer.SplitIntoTokens(content)
}

func (c *Converter) convertLargeDocument(ctx context.Context, content string) ([]map[string]interface{}, error) {
	logger.Info("Starting conversion of large document")
	segments := c.splitContent(content)
	logger.Debug(fmt.Sprintf("Document split into %d segments", len(segments)))

	var results []map[string]interface{}

	for i, segment := range segments {
		logger.Info(fmt.Sprintf("Processing segment %d of %d", i+1, len(segments)))
		jsonLD, newContext, err := c.convertSegmentWithContext(ctx, segment, c.analysisContext)
		if err != nil {
			logger.Error(fmt.Sprintf("Error converting segment %d: %v", i+1, err))
			return nil, fmt.Errorf("erreur lors de la conversion du segment %d : %w", i+1, err)
		}

		c.analysisContext = newContext

		jsonLD["segment"] = map[string]interface{}{
			"index": i + 1,
			"total": len(segments),
		}
		logger.Debug(fmt.Sprintf("Added segmentation metadata to segment %d", i+1))

		results = append(results, jsonLD)
		logger.Info(fmt.Sprintf("Segment %d processed successfully", i+1))
	}

	logger.Info(fmt.Sprintf("Large document conversion completed. Total segments processed: %d", len(segments)))
	return results, nil
}

func (c *Converter) convertSegmentWithContext(ctx context.Context, segment string, analysisContext *llm.AnalysisContext) (map[string]interface{}, *llm.AnalysisContext, error) {
	enrichedContent, newContext, err := c.llmClient.Analyze(ctx, segment, analysisContext)
	if err != nil {
		return nil, nil, fmt.Errorf("error enriching content with LLM: %w", err)
	}

	mainType, err := c.determineMainType(ctx, enrichedContent)
	if err != nil {
		logger.Warning(fmt.Sprintf("Error determining main type: %v. Falling back to 'Thing'", err))
		mainType = "Thing"
	}

	jsonLD, err := c.handleNestedStructures(ctx, enrichedContent, mainType)
	if err != nil {
		logger.Warning(fmt.Sprintf("Error handling nested structures: %v. Falling back to flat structure", err))
		jsonLD, err = c.extractProperties(ctx, enrichedContent, mainType)
		if err != nil {
			return nil, nil, fmt.Errorf("error extracting properties: %w", err)
		}
	}

	return jsonLD, newContext, nil
}
