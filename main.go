package main

import (
	"fmt"
	"io"
	"os"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func main() {
	// read file
	file, err := os.Open("test.md")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// read file content
	md, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	markdown := goldmark.New(
		goldmark.WithExtensions(
			meta.New(
				meta.WithStoresInDocument(),
			),
		),
	)

	doc := markdown.Parser().Parse(text.NewReader(md))

	book := NewBook()

	metadata := doc.OwnerDocument().Meta()
	if title, ok := metadata["title"]; ok {
		book.Title = title.(string)
	}
	if author, ok := metadata["author"]; ok {
		book.Author = author.(string)
	}
	if summary, ok := metadata["summary"]; ok {
		book.Summary = summary.(string)
	}

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n.(type) {
		case *ast.Heading:
			{
				heading := n.(*ast.Heading)
				if heading.Level != 1 {
					return ast.WalkContinue, nil
				}

				if book.Title == "" {
					book.Title = firstTextChild(&heading.BaseNode, md)
				}

				// confirm that the next node is a link
				link, ok := heading.FirstChild().(*ast.Link)
				if !ok {
					return ast.WalkContinue, nil
				}

				// get text of link
				linkText := firstTextChild(&link.BaseNode, md)
				linkHref := string(link.Destination)

				fmt.Println("Link Text:", linkText)
				fmt.Println("Link Href:", linkHref)
			}
		default:
			// fmt.Println("Not Heading")
		}

		return ast.WalkContinue, nil
	})

	fmt.Println("First Heading:", book.Title)
}

func firstTextChild(heading *ast.BaseNode, md []byte) string {
	var ret string
	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*ast.Text); ok {
			ret = string(text.Value(md))
			break
		}
	}
	return ret
}
