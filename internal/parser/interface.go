package parser

import (
	"io"
)

type Document struct {
	Content   string
	Metadata  map[string]string
	Structure []DocumentElement
}

type DocumentElement struct {
	Type     string // e.g., "paragraph", "heading", "list", etc.
	Content  string
	Children []DocumentElement
}

type Parser interface {
	Parse(r io.Reader) (*Document, error)
}
