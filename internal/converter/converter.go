package converter

import (
	"fmt"
	"sync"

	"github.com/chrlesur/json-ld-converter/internal/config"
	"github.com/chrlesur/json-ld-converter/internal/logger"
	"github.com/chrlesur/json-ld-converter/internal/parser"
	"github.com/chrlesur/json-ld-converter/pkg/tokenizer"
	"github.com/piprate/json-gold/ld"
)

// JSONLDConverter handles the conversion of parsed documents to JSON-LD format
type JSONLDConverter struct {
	schemaOrgContext string
	proc             *ld.JsonLdProcessor
	mutex            sync.Mutex
}

// NewJSONLDConverter initializes and returns a new JSONLDConverter instance
func NewJSONLDConverter() *JSONLDConverter {
	return &JSONLDConverter{
		schemaOrgContext: "https://schema.org",
		proc:             ld.NewJsonLdProcessor(),
	}
}

// Convert transforms a parsed document into a set of JSON-LD documents
func (c *JSONLDConverter) Convert(doc *parser.Document) ([]*ld.RDFDataset, error) {
	logger.Info("Starting JSON-LD conversion")

	c.mutex.Lock()
	defer c.mutex.Unlock()

	schemaOrgMap, err := c.mapToSchemaOrg(doc)
	if err != nil {
		return nil, fmt.Errorf("error mapping to Schema.org: %w", err)
	}

	jsonldMap := map[string]interface{}{
		"@context": c.schemaOrgContext,
		"@type":    "CreativeWork",
	}
	for k, v := range schemaOrgMap {
		jsonldMap[k] = v
	}

	segments, err := c.segmentJSONLD(jsonldMap)
	if err != nil {
		return nil, fmt.Errorf("error segmenting JSON-LD: %w", err)
	}

	var datasets []*ld.RDFDataset
	for _, segment := range segments {
		dataset, err := c.proc.ToRDF(segment, ld.NewJsonLdOptions(""))
		if err != nil {
			return nil, fmt.Errorf("error converting to RDF: %w", err)
		}
		datasets = append(datasets, dataset.(*ld.RDFDataset))
	}

	logger.Info(fmt.Sprintf("Conversion completed. Generated %d JSON-LD segments", len(datasets)))
	return datasets, nil
}

func (c *JSONLDConverter) mapToSchemaOrg(doc *parser.Document) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, element := range doc.Structure {
		switch element.Type {
		case "heading":
			result["headline"] = element.Content
		case "paragraph":
			if _, ok := result["text"]; !ok {
				result["text"] = element.Content
			} else {
				result["text"] = fmt.Sprintf("%s\n%s", result["text"], element.Content)
			}
		// Add more cases for other element types
		default:
			logger.Warning(fmt.Sprintf("Unhandled element type: %s", element.Type))
		}
	}

	// Add metadata
	for k, v := range doc.Metadata {
		result[k] = v
	}

	return result, nil
}

func (c *JSONLDConverter) segmentJSONLD(jsonld map[string]interface{}) ([]map[string]interface{}, error) {
	maxTokens := config.Get().Conversion.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4000 // Default value if not set in config
	}

	segments := []map[string]interface{}{}
	currentSegment := make(map[string]interface{})
	currentTokens := 0

	for k, v := range jsonld {
		valueTokens := tokenizer.CountTokens(fmt.Sprintf("%v", v))
		if currentTokens+valueTokens > maxTokens {
			if len(currentSegment) > 0 {
				segments = append(segments, currentSegment)
				currentSegment = make(map[string]interface{})
				currentTokens = 0
			}
		}
		currentSegment[k] = v
		currentTokens += valueTokens
	}

	if len(currentSegment) > 0 {
		segments = append(segments, currentSegment)
	}

	return segments, nil
}
