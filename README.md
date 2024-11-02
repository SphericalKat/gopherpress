# gopherpress

A simple tool to generate an epub book from a (properly) formatted markdown file.

## Installation

You can install `gopherpress` using `go install`:

```bash
go install github.com/sphericalkat/gopherpress@latest
```

## Description

`gopherpress` is a command-line tool with very few options. It reads a markdown file and generates an epub file. The markdown file must be formatted in a specific way. The first three lines of the file should be a YAML front matter block.

```
NAME:
   gopherpress - Turn HTML into an ebook using Markdown

USAGE:
   gopherpress [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --input value, -i value   Input file (Markdown)
   --output value, -o value  Output file (epub) (default: "output.epub")
   --help, -h                show hel
```

Given this input:

```
---
title: Crafting Interpreters
author: Robert Nystrom
summary: This book teaches you everything you need to know to implement a full-featured, efficient scripting language. You'll learn both high-level concepts around parsing and semantics and gritty details like bytecode representation and garbage collection. Your brain will light up with new ideas, and your hands will get dirty and calloused.
---

# Crafting Interpreters

# [Cover](https://craftinginterpreters.com/image/header.png)

# [Chapter 1: The Lox Language](https://craftinginterpreters.com/the-lox-language.html)

# [Chapter 2: Scanning](https://craftinginterpreters.com/scanning.html)

```

It will generate an epub file with the title "Crafting Interpreters" and the author "Robert Nystrom". The summary will be displayed on the cover page. The chapters will be displayed in the order they appear in the markdown file.
