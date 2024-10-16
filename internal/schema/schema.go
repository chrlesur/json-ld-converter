package schema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/chrlesur/json-ld-converter/internal/logger"
)

type SchemaType struct {
	ID           string            `json:"@id"`
	Label        string            `json:"rdfs:label"`
	Comment      string            `json:"rdfs:comment"`
	Properties   []string          `json:"properties,omitempty"`
	SubClassOf   []string          `json:"subClassOf,omitempty"`
	IsPartOf     string            `json:"isPartOf"`
	Source       string            `json:"source"`
	Enumerations map[string]string `json:"enumerations,omitempty"`
}

type SchemaProperty struct {
	ID             string   `json:"@id"`
	Label          string   `json:"rdfs:label"`
	Comment        string   `json:"rdfs:comment"`
	DomainIncludes []string `json:"domainIncludes,omitempty"`
	RangeIncludes  []string `json:"rangeIncludes,omitempty"`
	IsPartOf       string   `json:"isPartOf"`
	Source         string   `json:"source"`
}

type SchemaOrg struct {
	Types      map[string]SchemaType
	Properties map[string]SchemaProperty
}

func LoadSchemaOrg(filePath string) (*SchemaOrg, error) {
	logger.Debug(fmt.Sprintf("Loading Schema.org from file: %s", filePath))

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading Schema.org file: %w", err)
	}

	var jsonldSchema map[string]interface{}
	if err := json.Unmarshal(data, &jsonldSchema); err != nil {
		return nil, fmt.Errorf("error unmarshaling Schema.org data: %w", err)
	}

	schema := &SchemaOrg{
		Types:      make(map[string]SchemaType),
		Properties: make(map[string]SchemaProperty),
	}

	graph, ok := jsonldSchema["@graph"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("@graph element not found or not an array")
	}

	for _, item := range graph {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		id, ok := itemMap["@id"].(string)
		if !ok {
			continue
		}

		itemType, ok := itemMap["@type"].(string)
		if !ok {
			continue
		}

		switch itemType {
		case "rdfs:Class":
			schemaType := SchemaType{
				ID:      id,
				Label:   getStringValue(itemMap, "rdfs:label"),
				Comment: getStringValue(itemMap, "rdfs:comment"),
			}
			schema.Types[id] = schemaType
			logger.Debug(fmt.Sprintf("Loaded schema type: %s", id))
		case "rdf:Property":
			schemaProperty := SchemaProperty{
				ID:      id,
				Label:   getStringValue(itemMap, "rdfs:label"),
				Comment: getStringValue(itemMap, "rdfs:comment"),
			}
			schema.Properties[id] = schemaProperty
			logger.Debug(fmt.Sprintf("Loaded schema property: %s", id))
		default:
			schemaType := SchemaType{
				ID:      id,
				Label:   getStringValue(itemMap, "rdfs:label"),
				Comment: getStringValue(itemMap, "rdfs:comment"),
			}
			schema.Types[id] = schemaType
			logger.Debug(fmt.Sprintf("Loaded schema type: %s (type: %s)", id, itemType))
		}
	}

	if len(schema.Types) == 0 {
		return nil, fmt.Errorf("no types loaded from Schema.org")
	}

	logger.Info(fmt.Sprintf("Loaded %d types and %d properties from Schema.org", len(schema.Types), len(schema.Properties)))

	return schema, nil
}

func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func (s *SchemaOrg) GetType(typeName string) (SchemaType, bool) {
	// Try with "schema:" prefix
	if t, ok := s.Types["schema:"+typeName]; ok {
		logger.Debug(fmt.Sprintf("Found type with schema: prefix: %s", typeName))
		return t, true
	}
	// Try without prefix
	if t, ok := s.Types[typeName]; ok {
		logger.Debug(fmt.Sprintf("Found type without prefix: %s", typeName))
		return t, true
	}
	// Try with lowercase
	lowercaseTypeName := strings.ToLower(typeName)
	for key, value := range s.Types {
		if strings.ToLower(key) == lowercaseTypeName {
			logger.Debug(fmt.Sprintf("Found type with case-insensitive match: %s", key))
			return value, true
		}
	}
	logger.Warning(fmt.Sprintf("Type not found: %s", typeName))
	return SchemaType{}, false
}

func (s *SchemaOrg) GetProperty(propertyName string) (SchemaProperty, bool) {
	// Similar implementation as GetType, but for properties
	if p, ok := s.Properties["schema:"+propertyName]; ok {
		logger.Debug(fmt.Sprintf("Found property with schema: prefix: %s", propertyName))
		return p, true
	}
	if p, ok := s.Properties[propertyName]; ok {
		logger.Debug(fmt.Sprintf("Found property without prefix: %s", propertyName))
		return p, true
	}
	lowercasePropertyName := strings.ToLower(propertyName)
	for key, value := range s.Properties {
		if strings.ToLower(key) == lowercasePropertyName {
			logger.Debug(fmt.Sprintf("Found property with case-insensitive match: %s", key))
			return value, true
		}
	}
	logger.Warning(fmt.Sprintf("Property not found: %s", propertyName))
	return SchemaProperty{}, false
}

func (s *SchemaOrg) SuggestProperties(typeName string, content string) []string {
	schemaType, ok := s.GetType(typeName)
	if !ok {
		logger.Warning(fmt.Sprintf("Cannot suggest properties for unknown type: %s", typeName))
		return nil
	}

	var suggestedProperties []string
	for _, propName := range schemaType.Properties {
		prop, ok := s.GetProperty(strings.TrimPrefix(propName, "schema:"))
		if ok && strings.Contains(strings.ToLower(content), strings.ToLower(prop.Label)) {
			suggestedProperties = append(suggestedProperties, prop.ID)
			logger.Debug(fmt.Sprintf("Suggested property for %s: %s", typeName, prop.ID))
		}
	}

	logger.Info(fmt.Sprintf("Suggested %d properties for type %s", len(suggestedProperties), typeName))
	return suggestedProperties
}
