package parser

import (
	"strings"
	"testing"
)

func TestTextParser(t *testing.T) {
	input := "This is a test.\nThis is another line."
	parser := NewTextParser()
	doc, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to parse text: %v", err)
	}
	if doc.Content != input+"\n" {
		t.Errorf("Expected content %q, got %q", input+"\n", doc.Content)
	}
	if len(doc.Structure) != 2 {
		t.Errorf("Expected 2 paragraphs, got %d", len(doc.Structure))
	}
}

func TestMarkdownParser(t *testing.T) {
	input := "# Heading\n\nThis is a paragraph."
	parser := NewMarkdownParser()
	doc, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to parse markdown: %v", err)
	}
	if len(doc.Structure) != 2 {
		t.Errorf("Expected 2 elements (heading and paragraph), got %d", len(doc.Structure))
	}
	if doc.Structure[0].Type != "heading" || doc.Structure[1].Type != "paragraph" {
		t.Errorf("Unexpected structure types")
	}
}

// Add similar tests for PDF and HTML parsers
