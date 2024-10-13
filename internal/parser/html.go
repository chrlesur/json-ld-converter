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
