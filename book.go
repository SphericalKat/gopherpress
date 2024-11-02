package main

type Chapter struct {
	Title    string
	BodyHTML string
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

func (b *Book) AddChapter(title, body string) {
	b.Chapters = append(b.Chapters, Chapter{
		Title:    title,
		BodyHTML: body,
	})
}
