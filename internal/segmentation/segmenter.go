package segmentation

import (
	"strings"

	"github.com/chrlesur/json-ld-converter/internal/parser"
	"github.com/chrlesur/json-ld-converter/pkg/tokenizer"
)

type Segment struct {
	Content  string
	Metadata map[string]string
}

type Segmenter struct {
	maxTokens       int
	targetBatchSize int
}

func NewSegmenter(maxTokens, targetBatchSize int) *Segmenter {
	return &Segmenter{
		maxTokens:       maxTokens,
		targetBatchSize: targetBatchSize,
	}
}

func (s *Segmenter) Segment(doc *parser.Document) ([]Segment, error) {
	var segments []Segment
	var currentSegment strings.Builder
	currentTokens := 0

	for _, element := range doc.Structure {
		elementTokens := tokenizer.CountTokens(element.Content)

		if currentTokens+elementTokens > s.maxTokens {
			if currentSegment.Len() > 0 {
				segments = append(segments, Segment{
					Content:  currentSegment.String(),
					Metadata: make(map[string]string),
				})
				currentSegment.Reset()
				currentTokens = 0
			}
		}

		if elementTokens > s.maxTokens {
			// Si l'élément est trop grand, le diviser en sous-segments
			subSegments := s.splitLargeElement(element)
			segments = append(segments, subSegments...)
		} else {
			currentSegment.WriteString(element.Content)
			currentSegment.WriteString("\n")
			currentTokens += elementTokens

			if currentTokens >= s.targetBatchSize {
				segments = append(segments, Segment{
					Content:  currentSegment.String(),
					Metadata: make(map[string]string),
				})
				currentSegment.Reset()
				currentTokens = 0
			}
		}
	}

	if currentSegment.Len() > 0 {
		segments = append(segments, Segment{
			Content:  currentSegment.String(),
			Metadata: make(map[string]string),
		})
	}

	return segments, nil
}

func (s *Segmenter) splitLargeElement(element parser.DocumentElement) []Segment {
	var segments []Segment
	content := element.Content
	for len(content) > 0 {
		tokenCount := 0
		var segmentBuilder strings.Builder
		words := strings.Fields(content)

		for _, word := range words {
			wordTokens := tokenizer.CountTokens(word)
			if tokenCount+wordTokens > s.maxTokens {
				break
			}
			segmentBuilder.WriteString(word)
			segmentBuilder.WriteString(" ")
			tokenCount += wordTokens
		}

		segments = append(segments, Segment{
			Content:  segmentBuilder.String(),
			Metadata: map[string]string{"type": element.Type},
		})

		content = strings.TrimSpace(content[len(segmentBuilder.String()):])
	}

	return segments
}
