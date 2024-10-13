package converter

import (
	"reflect"
	"testing"

	"github.com/chrlesur/json-ld-converter/internal/parser"
)

func TestNewJSONLDConverter(t *testing.T) {
	converter := NewJSONLDConverter()
	if converter == nil {
		t.Error("NewJSONLDConverter returned nil")
	}
	if converter.schemaOrgContext != "https://schema.org" {
		t.Errorf("Expected schemaOrgContext to be 'https://schema.org', got '%s'", converter.schemaOrgContext)
	}
	if converter.proc == nil {
		t.Error("JsonLdProcessor is nil")
	}
}

func TestMapToSchemaOrg(t *testing.T) {
	converter := NewJSONLDConverter()
	doc := &parser.Document{
		Structure: []parser.DocumentElement{
			{Type: "heading", Content: "Test Heading"},
			{Type: "paragraph", Content: "Test paragraph"},
		},
		Metadata: map[string]string{"author": "Test Author"},
	}

	result, err := converter.mapToSchemaOrg(doc)
	if err != nil {
		t.Fatalf("mapToSchemaOrg returned an error: %v", err)
	}

	expected := map[string]interface{}{
		"headline": "Test Heading",
		"text":     "Test paragraph",
		"author":   "Test Author",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mapToSchemaOrg returned %v, expected %v", result, expected)
	}
}

func TestSegmentJSONLD(t *testing.T) {
	converter := NewJSONLDConverter()
	input := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "CreativeWork",
		"text":     "This is a very long text that should be split into multiple segments.",
	}

	segments, err := converter.segmentJSONLD(input)
	if err != nil {
		t.Fatalf("segmentJSONLD returned an error: %v", err)
	}

	if len(segments) != 1 {
		t.Errorf("Expected 1 segment, got %d", len(segments))
	}

	for _, segment := range segments {
		if segment["@context"] != "https://schema.org" {
			t.Errorf("Segment is missing @context")
		}
		if segment["@type"] != "CreativeWork" {
			t.Errorf("Segment is missing @type")
		}
		if _, ok := segment["text"]; !ok {
			t.Errorf("Segment is missing text")
		}
	}
}

func TestConvert(t *testing.T) {
	converter := NewJSONLDConverter()
	doc := &parser.Document{
		Structure: []parser.DocumentElement{
			{Type: "heading", Content: "Test Document"},
			{Type: "paragraph", Content: "This is a test paragraph."},
		},
		Metadata: map[string]string{"author": "Test Author"},
	}

	datasets, err := converter.Convert(doc)
	if err != nil {
		t.Fatalf("Convert returned an error: %v", err)
	}

	if len(datasets) != 1 {
		t.Errorf("Expected 1 dataset, got %d", len(datasets))
	}

	// Add more specific checks for the content of the datasets here
}
