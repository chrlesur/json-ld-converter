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