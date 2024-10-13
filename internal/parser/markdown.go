package parser

import (
	"io"

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
