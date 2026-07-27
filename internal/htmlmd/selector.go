package htmlmd

import (
	"strings"

	"golang.org/x/net/html"
)

// Selector is a minimal tag/class matcher used to identify elements that
// should be dropped from the output entirely (icon-font glyphs, badges,
// etc.) — the kind of thing that has no useful Markdown representation.
//
// Accepted string forms:
//
//	span              any <span> element
//	.material-icons   any element with class "material-icons"
//	i.fa              <i> elements with class "fa"
//	.fa-*             any element with a class starting with "fa-"
type Selector struct {
	Tag         string
	Class       string
	ClassPrefix bool
}

// ParseSelector parses a single selector string. Unrecognized input (empty
// string) yields a Selector that matches nothing.
func ParseSelector(s string) Selector {
	s = strings.TrimSpace(s)
	if s == "" {
		return Selector{}
	}

	tag, class, hasClass := strings.Cut(s, ".")
	sel := Selector{Tag: tag}
	if hasClass {
		if strings.HasSuffix(class, "*") {
			sel.Class = strings.TrimSuffix(class, "*")
			sel.ClassPrefix = true
		} else {
			sel.Class = class
		}
	}
	return sel
}

// ParseSelectors splits a comma-separated selector list.
func ParseSelectors(s string) []Selector {
	var out []Selector
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, ParseSelector(part))
	}
	return out
}

// Matches reports whether n satisfies the selector.
func (sel Selector) Matches(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if sel.Tag != "" && n.Data != sel.Tag {
		return false
	}
	if sel.Class == "" {
		return true
	}
	for _, c := range classList(n) {
		if sel.ClassPrefix {
			if strings.HasPrefix(c, sel.Class) {
				return true
			}
		} else if c == sel.Class {
			return true
		}
	}
	return false
}

func classList(n *html.Node) []string {
	val, _ := attr(n, "class")
	return strings.Fields(val)
}

// DefaultIconSelectors match the common icon-font conventions (Google
// Material Icons/Symbols, Font Awesome, Bootstrap Icons, Octicons). These
// elements typically carry either a private-use glyph or a ligature word
// (e.g. "arrow_back") as their text content — neither is meaningful in
// Markdown, so by default such elements are dropped entirely.
var DefaultIconSelectors = []Selector{
	{Class: "material-icons"},
	{Class: "material-icons-outlined"},
	{Class: "material-icons-round"},
	{Class: "material-icons-sharp"},
	{Class: "material-icons-two-tone"},
	{Class: "material-symbols-outlined"},
	{Class: "material-symbols-rounded"},
	{Class: "material-symbols-sharp"},
	{Tag: "i", Class: "fa"},
	{Tag: "i", Class: "fas"},
	{Tag: "i", Class: "far"},
	{Tag: "i", Class: "fal"},
	{Tag: "i", Class: "fab"},
	{Tag: "i", Class: "fad"},
	{Class: "fa-", ClassPrefix: true},
	{Class: "glyphicon", ClassPrefix: true},
	{Class: "bi-", ClassPrefix: true},
	{Class: "octicon", ClassPrefix: true},
}
