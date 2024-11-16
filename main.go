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

	"github.com/fatih/color"
	"github.com/go-shiori/go-epub"
	readability "github.com/go-shiori/go-readability"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "gopherpress",
		Usage: "Turn HTML into an ebook using Markdown",
		Action: func(c *cli.Context) error {
			input := c.String("input")
			output := c.String("output")

			run(c, input, output)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "Input file (Markdown)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "Output file (epub)",
				Value:    "output.epub",
				Required: false,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(color.RedString("error: %s", err))
	}
}

func run(_ *cli.Context, input string, output string) {
	// read file
	file, err := os.Open(input)
	if err != nil {
		fmt.Println(color.RedString("error: %s", err))
		os.Exit(1)
	}
	defer file.Close()

	// read file content
	md, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(color.RedString("error: %s", err))
		os.Exit(1)
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
	} else {
		fmt.Println(color.YellowString("warning: no title found, gopherpress will try to extract it from the first heading"))
	}

	if author, ok := metadata["author"]; ok {
		book.Author = author.(string)
	} else {
		fmt.Println(color.YellowString("warning: no author found"))
	}

	if summary, ok := metadata["summary"]; ok {
		book.Summary = summary.(string)
	} else {
		fmt.Println(color.YellowString("warning: no summary found"))
	}

	if coverImg, ok := metadata["cover"]; ok {
		book.CoverImg = coverImg.(string)
	} else {
		fmt.Println(color.YellowString("warning: no cover image found"))
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

				// extract text using readability
				article, err := readability.FromReader(strings.NewReader(string(linkContent)), nil)
				if err != nil {
					fmt.Println(err)
					return ast.WalkStop, nil
				}

				book.AddChapter(linkText, article.Content, linkHref)
			}
		}

		return ast.WalkContinue, nil
	})

	e, err := epub.NewEpub(book.Title)
	if err != nil {
		fmt.Println(color.RedString("error: %s", err))
		os.Exit(1)
	}

	e.SetAuthor(book.Author)

	for _, chapter := range book.Chapters {
		chapter.PreProcessHTML(chapter.LinkHref)
		_, err := e.AddSection(chapter.BodyHTML, chapter.Title, "", "")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	e.EmbedImages()

	// Add cover image
	if len(book.CoverImg) > 0 {
		coverImagePath, err := e.AddImage(book.CoverImg, "")
		if err != nil {
			fmt.Println(color.RedString("error: %s", err))
			os.Exit(1)
		}

		e.SetCover(coverImagePath, "")
	}

	err = e.Write(output)
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
