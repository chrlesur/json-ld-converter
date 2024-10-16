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

	// Créer un io.ReaderAt à partir du contenu
	contentReader := strings.NewReader(string(content))
	size := int64(len(content))

	reader, err := pdf.NewReader(contentReader, size)
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
