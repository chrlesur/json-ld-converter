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

func SegmentDocument(doc *parser.Document, maxTokens int) ([]Segment, error) {
    var segments []Segment
    var currentSegment strings.Builder
    currentTokens := 0

    for _, element := range doc.Structure {
        elementTokens := tokenizer.CountTokens(element.Content)

        if elementTokens > maxTokens {
            // Si l'élément est trop grand, le diviser en sous-segments
            subSegments := splitLargeElement(element, maxTokens)
            segments = append(segments, subSegments...)
        } else if currentTokens+elementTokens > maxTokens {
            if currentSegment.Len() > 0 {
                segments = append(segments, Segment{
                    Content:  currentSegment.String(),
                    Metadata: make(map[string]string),
                })
                currentSegment.Reset()
                currentTokens = 0
            }
            currentSegment.WriteString(element.Content)
            currentSegment.WriteString("\n")
            currentTokens += elementTokens
        } else {
            currentSegment.WriteString(element.Content)
            currentSegment.WriteString("\n")
            currentTokens += elementTokens
        }

        if currentTokens >= maxTokens {
            segments = append(segments, Segment{
                Content:  currentSegment.String(),
                Metadata: make(map[string]string),
            })
            currentSegment.Reset()
            currentTokens = 0
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

func splitLargeElement(element parser.DocumentElement, maxTokens int) []Segment {
    var segments []Segment
    content := element.Content
    for len(content) > 0 {
        tokenCount := 0
        var segmentBuilder strings.Builder
        words := strings.Fields(content)

        for _, word := range words {
            wordTokens := tokenizer.CountTokens(word)
            if tokenCount+wordTokens > maxTokens {
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
