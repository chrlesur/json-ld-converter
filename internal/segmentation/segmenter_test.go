package segmentation

import (
	"strings"
	"testing"

	"github.com/chrlesur/json-ld-converter/internal/parser"
	"github.com/chrlesur/json-ld-converter/pkg/tokenizer"
)

func TestSegmenter(t *testing.T) {
	doc := &parser.Document{
		Structure: []parser.DocumentElement{
			{Type: "heading", Content: "Title"},
			{Type: "paragraph", Content: "This is a long paragraph that should be split into multiple segments. " + strings.Repeat("More content. ", 100)},
			{Type: "list", Content: "Item 1\nItem 2\nItem 3"},
		},
	}

	segmenter := NewSegmenter(1000, 500)
	segments, err := segmenter.Segment(doc)

	if err != nil {
		t.Fatalf("Segmentation failed: %v", err)
	}

	if len(segments) < 2 {
		t.Errorf("Expected multiple segments, got %d", len(segments))
	}

	for i, segment := range segments {
		tokens := tokenizer.CountTokens(segment.Content)
		if tokens > 1000 {
			t.Errorf("Segment %d exceeds max tokens: %d", i, tokens)
		}
	}
}
