package jsonld

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

	logger.Debug(fmt.Sprintf("Starting conversion of document with content: %s", doc.Content))

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
	logger.Debug(fmt.Sprintf("Enriched content: %s", enrichedContent))

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
	logger.Debug(fmt.Sprintf("Nested content: %+v", nestedContent))
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

	schemaType, ok := c.schemaOrg.GetType(mainType)
	if !ok {
		return nil, fmt.Errorf("type non trouvé dans le schéma : %s", mainType)
	}

	prompt := fmt.Sprintf(`Analysez le contenu suivant et extrayez les propriétés pertinentes pour un objet de type '%s' selon le schéma Schema.org. Retournez UNIQUEMENT un objet JSON valide, sans texte supplémentaire avant ou après.

Contenu à analyser :
%s

Propriétés possibles pour le type '%s' :
%s

Instructions spéciales :
- Utilisez "mentions" pour lister les personnages ou entités importantes, incluant leurs actions principales sous forme de texte.
- Incluez "events" comme un tableau d'objets, chacun avec un "name" et une "description" détaillée de l'événement.
- Utilisez "description" pour fournir un résumé détaillé incluant les actions et relations entre les personnes.
- Pour "keywords", fournissez une liste de mots-clés pertinents, y compris des verbes d'action.
- Incluez "datePublished" au format YYYY-MM-DD si une date de publication est mentionnée.
- Incluez "author" avec le nom de l'auteur si mentionné.
- Incluez "genre" si le genre de l'œuvre est spécifié.

N'incluez PAS les propriétés "@context" et "@type" dans votre réponse.`, mainType, content, mainType, strings.Join(schemaType.Properties, ", "))

	response, _, err := c.llmClient.Analyze(ctx, prompt, &llm.AnalysisContext{})
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'analyse LLM : %w", err)
	}

	logger.Debug(fmt.Sprintf("LLM response for properties: %s", response))

	jsonStartIndex := strings.Index(response, "{")
	if jsonStartIndex == -1 {
		return nil, fmt.Errorf("aucun JSON trouvé dans la réponse LLM")
	}

	jsonResponse := response[jsonStartIndex:]

	var extractedProperties map[string]interface{}
	err = json.Unmarshal([]byte(jsonResponse), &extractedProperties)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du parsing de la réponse JSON : %w", err)
	}

	properties := make(map[string]interface{})
	for key, value := range extractedProperties {
		switch key {
		case "mentions":
			if persons, ok := value.([]interface{}); ok {
				mentions := make([]map[string]interface{}, 0)
				for _, person := range persons {
					if p, ok := person.(map[string]interface{}); ok {
						mention := map[string]interface{}{
							"@type": "Person",
							"name":  p["name"],
						}
						if action, ok := p["action"].(string); ok && action != "" {
							mention["description"] = action
						}
						mentions = append(mentions, mention)
					}
				}
				if len(mentions) > 0 {
					properties["mentions"] = mentions
				}
			}
		case "events":
			if events, ok := value.([]interface{}); ok {
				eventsList := make([]map[string]interface{}, 0)
				for _, event := range events {
					if e, ok := event.(map[string]interface{}); ok {
						if name, ok := e["name"].(string); ok && name != "" {
							eventItem := map[string]interface{}{
								"@type": "Event",
								"name":  name,
							}
							if desc, ok := e["description"].(string); ok && desc != "" {
								eventItem["description"] = desc
							}
							eventsList = append(eventsList, eventItem)
						}
					}
				}
				if len(eventsList) > 0 {
					properties["events"] = eventsList
				}
			}
		case "keywords":
			if keywords, ok := value.([]interface{}); ok {
				keywordStrings := make([]string, 0)
				for _, kw := range keywords {
					if kwStr, ok := kw.(string); ok && kwStr != "" {
						keywordStrings = append(keywordStrings, kwStr)
					}
				}
				if len(keywordStrings) > 0 {
					properties["keywords"] = strings.Join(keywordStrings, ", ")
				}
			}
		case "datePublished":
			if dateStr, ok := value.(string); ok && isValidDate(dateStr) {
				properties["datePublished"] = dateStr
			}
		case "author":
			if authorName, ok := value.(string); ok && authorName != "" {
				properties["author"] = map[string]interface{}{
					"@type": "Person",
					"name":  authorName,
				}
			}
		default:
			if str, ok := value.(string); ok && str != "" {
				properties[key] = str
			} else {
				properties[key] = value
			}
		}
	}

	logger.Debug(fmt.Sprintf("All extracted properties: %+v", properties))

	return properties, nil
}

// Fonction utilitaire pour valider le format de date
func isValidDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
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

	logger.Debug(fmt.Sprintf("Extracted properties: %+v", properties))

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

	logger.Debug(fmt.Sprintf("Final jsonLD structure: %+v", jsonLD))
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
