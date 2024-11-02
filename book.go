package main

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type Chapter struct {
	Title    string
	BodyHTML string
	LinkHref string
}

type Book struct {
	Title    string
	Author   string
	Summary  string
	Chapters []Chapter
	CoverImg string
}

func NewBook() *Book {
	return &Book{
		Chapters: make([]Chapter, 0),
	}
}

func (b *Book) AddChapter(title, body, linkHref string) {
	b.Chapters = append(b.Chapters, Chapter{
		Title:    title,
		BodyHTML: body,
		LinkHref: linkHref,
	})
}

func (c *Chapter) EmbedImgUrls(baseURL string) error {
	base, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	// remove path from base url
	base.Path = ""

	// Embed image urls in the chapter
	doc, err := html.Parse(strings.NewReader(c.BodyHTML))
	if err != nil {
		return err
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for i, attr := range n.Attr {
				if attr.Key == "src" {
					relativeURL, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}
					absoluteURL := base.ResolveReference(relativeURL).String()
					n.Attr[i].Val = absoluteURL
				}
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			traverse(ch)
		}
	}

	traverse(doc)
	var buf strings.Builder
	err = html.Render(&buf, doc)
	if err != nil {
		return err
	}

	c.BodyHTML = buf.String()

	return nil
}
