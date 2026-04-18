package renderer

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Sections holds the rendered HTML for each H2 section in a project markdown file.
type Sections struct {
	Problem      string
	Architecture string
	Decisions    string
	Tradeoffs    string
	Failures     string
	Improvements string
}

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
}

// ParseProjectMarkdown renders the full markdown to HTML, then extracts
// each H2 section by heading anchor ID.
func ParseProjectMarkdown(source []byte) (*Sections, error) {
	var buf bytes.Buffer
	ctx := parser.NewContext()

	if err := md.Convert(source, &buf, parser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("goldmark convert: %w", err)
	}

	rendered := buf.String()

	return &Sections{
		Problem:      extractSection(rendered, "problem"),
		Architecture: extractSection(rendered, "architecture"),
		Decisions:    extractSection(rendered, "design-decisions"),
		Tradeoffs:    extractSection(rendered, "trade-offs"),
		Failures:     extractSection(rendered, "failure-handling"),
		Improvements: extractSection(rendered, "improvements"),
	}, nil
}

// extractSection pulls out HTML content between two H2 headings.
// It looks for <h2 id="{id}"> and extracts everything until the next <h2.
func extractSection(html, id string) string {
	// Find opening tag with this id
	openTag := fmt.Sprintf(`id="%s"`, id)
	start := strings.Index(html, openTag)
	if start == -1 {
		return ""
	}

	// Find the end of the opening h2 tag
	closingAngle := strings.Index(html[start:], ">")
	if closingAngle == -1 {
		return ""
	}
	// Move past the </h2> closing tag
	afterHeading := start + closingAngle + 1
	closeH2 := strings.Index(html[afterHeading:], "</h2>")
	if closeH2 == -1 {
		return ""
	}
	contentStart := afterHeading + closeH2 + len("</h2>")

	// Find the next <h2 to end the section
	rest := html[contentStart:]
	nextH2 := strings.Index(rest, "<h2")
	if nextH2 == -1 {
		return strings.TrimSpace(rest)
	}

	return strings.TrimSpace(rest[:nextH2])
}