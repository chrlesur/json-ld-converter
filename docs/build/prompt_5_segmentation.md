# Implémentation du système de segmentation pour le Convertisseur JSON-LD

Objectif : Créer un système de segmentation capable de diviser de grands documents (jusqu'à 120 000 tokens) en segments gérables tout en préservant le contexte, avec une limite de 4 000 tokens par segment de sortie JSON-LD.

## Tâches :

1. Dans le répertoire `internal/segmentation`, créez un fichier `segmenter.go` avec le contenu suivant :

```go
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
	maxTokens      int
	targetBatchSize int
}

func NewSegmenter(maxTokens, targetBatchSize int) *Segmenter {
	return &Segmenter{
		maxTokens:      maxTokens,
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
```

2. Créez un fichier `segmenter_test.go` dans le même répertoire pour tester le segmenter :

```go
package segmentation

import (
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
```

3. Dans le répertoire `pkg`, créez un package `tokenizer` avec un fichier `tokenizer.go` :

```go
package tokenizer

import (
	"strings"
	"unicode"
)

func CountTokens(text string) int {
	return len(strings.Fields(text))
}

func SplitIntoTokens(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
}
```

4. Modifiez le fichier `internal/config/config.go` pour inclure les paramètres de segmentation :

```go
type Config struct {
	// ... autres champs existants ...
	Segmentation struct {
		MaxTokens      int `yaml:"max_tokens"`
		TargetBatchSize int `yaml:"target_batch_size"`
	} `yaml:"segmentation"`
}
```

5. Mettez à jour le fichier de configuration `config.yaml` pour inclure les nouveaux paramètres :

```yaml
# ... autres configurations existantes ...
segmentation:
  max_tokens: 4000
  target_batch_size: 1000
```

6. Dans le fichier principal de votre application (par exemple, `cmd/converter/main.go`), ajoutez le code pour utiliser le segmenter :

```go
package main

import (
	"log"

	"github.com/chrlesur/json-ld-converter/internal/config"
	"github.com/chrlesur/json-ld-converter/internal/parser"
	"github.com/chrlesur/json-ld-converter/internal/segmentation"
)

func main() {
	// ... code existant pour charger la configuration ...

	cfg := config.Get()

	// Exemple d'utilisation du segmenter
	doc, err := loadDocument("path/to/document") // Implémentez cette fonction
	if err != nil {
		log.Fatalf("Failed to load document: %v", err)
	}

	segmenter := segmentation.NewSegmenter(
		cfg.Segmentation.MaxTokens,
		cfg.Segmentation.TargetBatchSize,
	)

	segments, err := segmenter.Segment(doc)
	if err != nil {
		log.Fatalf("Failed to segment document: %v", err)
	}

	for i, segment := range segments {
		log.Printf("Segment %d: %d tokens", i, tokenizer.CountTokens(segment.Content))
	}

	// ... code pour traiter les segments ...
}
```

## Utilisation du système de segmentation :

Pour utiliser le système de segmentation dans d'autres parties du projet, importez le package et utilisez-le comme suit :

```go
import (
	"github.com/chrlesur/json-ld-converter/internal/segmentation"
	"github.com/chrlesur/json-ld-converter/internal/parser"
)

func processDocument(doc *parser.Document) error {
	segmenter := segmentation.NewSegmenter(4000, 1000)
	segments, err := segmenter.Segment(doc)
	if err != nil {
		return err
	}

	for _, segment := range segments {
		// Traitez chaque segment ici
		// Par exemple, convertissez-le en JSON-LD
	}

	return nil
}
```

## Notes importantes :
- Le système de segmentation utilise une approche simple basée sur le comptage de tokens. Pour une tokenization plus précise, vous pourriez vouloir utiliser une bibliothèque spécialisée.
- La segmentation préserve la structure du document autant que possible, mais de très longs éléments peuvent être divisés.
- Assurez-vous de gérer correctement les métadonnées et le contexte lors du traitement des segments.
- Le système actuel ne gère pas les références croisées entre segments. Si c'est nécessaire pour votre cas d'utilisation, vous devrez implémenter un système de liaison entre segments.
- Testez le système avec différents types et tailles de documents pour vous assurer qu'il fonctionne correctement dans tous les cas.

Veuillez implémenter ce système de segmentation et effectuer les tests nécessaires. Une fois terminé, nous pourrons passer à l'étape suivante du développement du convertisseur JSON-LD.