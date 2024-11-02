package main

import (
	"strings"

	"golang.org/x/net/html"
)

func GetTextContent(n *html.Node) string {
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text += c.Data
		} else if c.Type == html.ElementNode && c.Data != "small" { // Ignore small tag
			text += GetTextContent(c)
		}
	}
	return strings.TrimSpace(text)
}
