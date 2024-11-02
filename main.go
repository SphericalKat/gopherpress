package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/go-shiori/go-epub"
	readability "github.com/go-shiori/go-readability"
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
				if !entering {
					return ast.WalkContinue, nil
				}

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

				// extract content from link
				var linkContent []byte
				if _, err := os.Stat(linkHref); !errors.Is(err, os.ErrNotExist) {
					linkContent, err = getFileContent(linkHref)
					if err != nil {
						fmt.Println(err)
						return ast.WalkStop, nil
					}
				} else {
					linkContent, err = getUrlContent(linkHref)
					if err != nil {
						fmt.Println(err)
						return ast.WalkStop, nil
					}
				}
				if err != nil {
					fmt.Println(err)
					return ast.WalkStop, nil
				}

				// if the link text is "cover", set the cover image
				// and stop walking the tree
				if strings.ToLower(string(linkText)) == "cover" {
					book.CoverImg = linkHref
					return ast.WalkContinue, nil
				}

				// extract text using readability
				article, err := readability.FromReader(strings.NewReader(string(linkContent)), nil)
				if err != nil {
					fmt.Println(err)
					return ast.WalkStop, nil
				}

				book.AddChapter(string(linkText), article.Content)
			}
		default:
			// fmt.Println("Not Heading")
		}

		return ast.WalkContinue, nil
	})

	fmt.Println("First Heading:", book.Title)

	e, err := epub.NewEpub(book.Title)
	if err != nil {
		fmt.Println(err)
		return
	}

	e.SetAuthor(book.Author)

	// Add cover image
	if len(book.CoverImg) > 0 {
		coverImagePath, err := e.AddImage(book.CoverImg, "")
		if err != nil {
			fmt.Println(err)
			return
		}

		e.SetCover(coverImagePath, "")
	}

	for _, chapter := range book.Chapters {
		_, err := e.AddSection(chapter.BodyHTML, chapter.Title, "", "")
		if err != nil {
			fmt.Println(err)
			return
		}
		e.EmbedImages()
	}

	err = e.Write("output.epub")
	if err != nil {
		fmt.Println(err)
		return
	}
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

func getUrlContent(url string) ([]byte, error) {
	var content []byte

	// fetch whatever is at the url
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// read the response body
	content, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func getFileContent(url string) ([]byte, error) {
	var content []byte

	file, err := os.Open(url)
	if err != nil {
		return nil, err
	}

	content, err = io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return content, nil
}
