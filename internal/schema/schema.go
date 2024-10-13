package schema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
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
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading Schema.org file: %w", err)
	}

	var rawSchema map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		return nil, fmt.Errorf("error unmarshaling Schema.org data: %w", err)
	}

	schema := &SchemaOrg{
		Types:      make(map[string]SchemaType),
		Properties: make(map[string]SchemaProperty),
	}

	for key, value := range rawSchema {
		if strings.HasPrefix(key, "schema:") {
			var schemaType SchemaType
			if err := json.Unmarshal(value, &schemaType); err == nil {
				schema.Types[key] = schemaType
			} else {
				var schemaProperty SchemaProperty
				if err := json.Unmarshal(value, &schemaProperty); err == nil {
					schema.Properties[key] = schemaProperty
				}
			}
		}
	}

	return schema, nil
}

func (s *SchemaOrg) GetType(typeName string) (SchemaType, bool) {
	t, ok := s.Types["schema:"+typeName]
	return t, ok
}

func (s *SchemaOrg) GetProperty(propertyName string) (SchemaProperty, bool) {
	p, ok := s.Properties["schema:"+propertyName]
	return p, ok
}

func (s *SchemaOrg) SuggestProperties(typeName string, content string) []string {
	schemaType, ok := s.GetType(typeName)
	if !ok {
		return nil
	}

	var suggestedProperties []string
	for _, propName := range schemaType.Properties {
		prop, ok := s.GetProperty(strings.TrimPrefix(propName, "schema:"))
		if ok && strings.Contains(strings.ToLower(content), strings.ToLower(prop.Label)) {
			suggestedProperties = append(suggestedProperties, prop.ID)
		}
	}

	return suggestedProperties
}
