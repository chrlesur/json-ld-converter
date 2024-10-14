# Implémentation des parseurs de documents pour le Convertisseur JSON-LD

Objectif : Développer des modules séparés pour l'analyse de documents texte, PDF, Markdown et HTML, avec une interface commune pour standardiser le processus d'extraction.

## Tâches :

1. Dans le répertoire `internal/parser`, créez un fichier `interface.go` avec le contenu suivant :

```go
package parser

import (
	"io"
)

type Document struct {
	Content     string
	Metadata    map[string]string
	Structure   []DocumentElement
}

type DocumentElement struct {
	Type     string // e.g., "paragraph", "heading", "list", etc.
	Content  string
	Children []DocumentElement
}

type Parser interface {
	Parse(r io.Reader) (*Document, error)
}
```

2. Créez un fichier `text.go` pour le parseur de texte brut :

```go
package parser

import (
	"bufio"
	"io"
	"strings"
)

type TextParser struct{}

func NewTextParser() *TextParser {
	return &TextParser{}
}

func (p *TextParser) Parse(r io.Reader) (*Document, error) {
	scanner := bufio.NewScanner(r)
	var content strings.Builder
	var structure []DocumentElement

	for scanner.Scan() {
		line := scanner.Text()
		content.WriteString(line + "\n")
		structure = append(structure, DocumentElement{
			Type:    "paragraph",
			Content: line,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &Document{
		Content:   content.String(),
		Metadata:  make(map[string]string),
		Structure: structure,
	}, nil
}
```

3. Créez un fichier `markdown.go` pour le parseur Markdown :

```go
package parser

import (
	"bufio"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MarkdownParser struct{}

func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

func (p *MarkdownParser) Parse(r io.Reader) (*Document, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	md := goldmark.New()
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	var structure []DocumentElement
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindHeading:
			heading := n.(*ast.Heading)
			structure = append(structure, DocumentElement{
				Type:    "heading",
				Content: string(heading.Text(content)),
			})
		case ast.KindParagraph:
			paragraph := n.(*ast.Paragraph)
			structure = append(structure, DocumentElement{
				Type:    "paragraph",
				Content: string(paragraph.Text(content)),
			})
		// Add more cases for other Markdown elements as needed
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, err
	}

	return &Document{
		Content:   string(content),
		Metadata:  make(map[string]string),
		Structure: structure,
	}, nil
}
```

4. Pour le parseur PDF, nous allons utiliser une bibliothèque externe. Ajoutez la dépendance :

```
go get github.com/ledongthuc/pdf
```

Puis créez un fichier `pdf.go` :

```go
package parser

import (
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

type PDFParser struct{}

func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

func (p *PDFParser) Parse(r io.Reader) (*Document, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	reader, err := pdf.NewReader(strings.NewReader(string(content)))
	if err != nil {
		return nil, err
	}

	var contentBuilder strings.Builder
	var structure []DocumentElement

	numPages := reader.NumPage()

	for pageIndex := 1; pageIndex <= numPages; pageIndex++ {
		page := reader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return nil, err
		}

		contentBuilder.WriteString(text)
		structure = append(structure, DocumentElement{
			Type:    "page",
			Content: text,
		})
	}

	return &Document{
		Content:   contentBuilder.String(),
		Metadata:  make(map[string]string),
		Structure: structure,
	}, nil
}
```

5. Pour le parseur HTML, nous utiliserons la bibliothèque standard `golang.org/x/net/html`. Créez un fichier `html.go` :

```go
package parser

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

type HTMLParser struct{}

func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

func (p *HTMLParser) Parse(r io.Reader) (*Document, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var contentBuilder strings.Builder
	var structure []DocumentElement

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			contentBuilder.WriteString(n.Data)
		}
		if n.Type == html.ElementNode {
			element := DocumentElement{
				Type: n.Data,
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
				if c.Type == html.TextNode {
					element.Content += c.Data
				}
			}
			structure = append(structure, element)
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(doc)

	return &Document{
		Content:   contentBuilder.String(),
		Metadata:  make(map[string]string),
		Structure: structure,
	}, nil
}
```

6. Créez un fichier `factory.go` pour faciliter la création des parseurs :

```go
package parser

import "fmt"

func NewParser(fileType string) (Parser, error) {
	switch fileType {
	case "text":
		return NewTextParser(), nil
	case "markdown":
		return NewMarkdownParser(), nil
	case "pdf":
		return NewPDFParser(), nil
	case "html":
		return NewHTMLParser(), nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}
```

7. Créez un fichier de test `parser_test.go` pour tester les parseurs :

```go
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
```

## Utilisation des parseurs :

Pour utiliser les parseurs dans d'autres parties du projet, importez le package et utilisez-le comme suit :

```go
import (
	"os"
	"github.com/chrlesur/json-ld-converter/internal/parser"
)

func processFile(filePath string, fileType string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	p, err := parser.NewParser(fileType)
	if err != nil {
		return err
	}

	doc, err := p.Parse(file)
	if err != nil {
		return err
	}

	// Utilisez doc.Content, doc.Metadata, et doc.Structure comme nécessaire
	return nil
}
```

## Notes importantes :
- Assurez-vous d'avoir installé toutes les dépendances nécessaires (`github.com/yuin/goldmark` pour Markdown, `github.com/ledongthuc/pdf` pour PDF).
- Le parseur PDF actuel extrait uniquement le texte brut. Pour une analyse plus détaillée, vous devrez peut-être utiliser une bibliothèque plus avancée ou implémenter une logique supplémentaire.
- Le parseur HTML actuel est basique. Vous pourriez vouloir l'améliorer pour extraire plus d'informations structurelles si nécessaire.
- Les métadonnées sont actuellement vides pour tous les parseurs. Vous devrez les remplir en fonction des spécificités de chaque format de document.
- Assurez-vous de gérer les erreurs et les cas limites dans votre code de production.

Veuillez implémenter ces parseurs de documents et effectuer les tests nécessaires. Une fois terminé, nous pourrons passer à l'étape suivante du développement du convertisseur JSON-LD.