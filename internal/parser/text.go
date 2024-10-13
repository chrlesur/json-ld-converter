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