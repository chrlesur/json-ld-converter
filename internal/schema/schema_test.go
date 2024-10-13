package schema

import (
	"testing"
)

func TestLoadSchemaOrg(t *testing.T) {
	schema, err := LoadSchemaOrg("testdata/schema.json")
	if err != nil {
		t.Fatalf("Failed to load Schema.org: %v", err)
	}

	if len(schema.Types) == 0 {
		t.Error("No types loaded from Schema.org")
	}

	if len(schema.Properties) == 0 {
		t.Error("No properties loaded from Schema.org")
	}

	// Test GetType
	person, ok := schema.GetType("Person")
	if !ok {
		t.Error("Failed to get Person type")
	} else if person.Label != "Person" {
		t.Errorf("Unexpected label for Person: %s", person.Label)
	}

	// Test GetProperty
	name, ok := schema.GetProperty("name")
	if !ok {
		t.Error("Failed to get name property")
	} else if name.Label != "name" {
		t.Errorf("Unexpected label for name property: %s", name.Label)
	}

	// Test SuggestProperties
	suggestedProps := schema.SuggestProperties("Person", "John Doe is 30 years old")
	if len(suggestedProps) == 0 {
		t.Error("No properties suggested for Person")
	}
}
