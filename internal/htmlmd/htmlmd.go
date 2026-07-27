// Package htmlmd converts an HTML document into clean Markdown, stripping
// non-content chrome (nav, header, footer, scripts, etc.) along the way.
package htmlmd

import (
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// noise holds tag names that never contribute content and are dropped
// entirely, along with their children.
var noise = map[string]bool{
	"script":   true,
	"style":    true,
	"noscript": true,
	"nav":      true,
	"header":   true,
	"footer":   true,
	"aside":    true,
	"iframe":   true,
	"svg":      true,
	"template": true,
	"button":   true,
	"form":     true,
}

// Options controls optional behavior of Convert.
type Options struct {
	// ExtraSelectors are additional elements to drop from the output,
	// on top of DefaultIconSelectors (unless DisableDefaultIcons is set).
	ExtraSelectors []Selector
	// DisableDefaultIcons turns off the built-in icon-font selectors
	// (Material Icons/Symbols, Font Awesome, Bootstrap Icons, Octicons).
	DisableDefaultIcons bool
}

// Convert parses raw HTML and returns cleaned Markdown text, using the
// default icon-font filters and no extra selectors.
func Convert(src []byte) (string, error) {
	return ConvertWithOptions(src, Options{})
}

// ConvertWithOptions parses raw HTML and returns cleaned Markdown text,
// dropping any element matched by opts' selectors along with default noise
// elements (script, style, nav, header, footer, aside, etc.).
func ConvertWithOptions(src []byte, opts Options) (string, error) {
	doc, err := html.Parse(strings.NewReader(string(src)))
	if err != nil {
		return "", err
	}

	body := findBody(doc)
	if body == nil {
		body = doc
	}

	var filters []Selector
	if !opts.DisableDefaultIcons {
		filters = append(filters, DefaultIconSelectors...)
	}
	filters = append(filters, opts.ExtraSelectors...)

	c := &converter{filters: filters}
	c.walkChildren(body)

	out := c.buf.String()
	out = collapseBlankLines(out)
	return strings.TrimSpace(out) + "\n", nil
}

func findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findBody(c); found != nil {
			return found
		}
	}
	return nil
}

type converter struct {
	buf         strings.Builder
	listDepth   int
	orderedIdx  []int // per-depth counter for <ol>
	atLineStart bool
	filters     []Selector
}

// dropped reports whether n should be excluded from the output entirely,
// per the built-in noise list or a configured filter selector.
func (c *converter) dropped(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if noise[n.Data] {
		return true
	}
	for _, sel := range c.filters {
		if sel.Matches(n) {
			return true
		}
	}
	return false
}

func (c *converter) writeString(s string) {
	c.buf.WriteString(s)
	if len(s) > 0 {
		c.atLineStart = strings.HasSuffix(s, "\n")
	}
}

// blockGap ensures the buffer ends with exactly one blank line, i.e. a
// paragraph break, before the next block-level element is written.
func (c *converter) blockGap() {
	s := c.buf.String()
	s = strings.TrimRight(s, " \t")
	if len(s) == 0 {
		return
	}
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	if !strings.HasSuffix(s, "\n\n") {
		s += "\n"
	}
	c.buf.Reset()
	c.buf.WriteString(s)
	c.atLineStart = true
}

func (c *converter) walkChildren(n *html.Node) {
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		c.walk(ch)
	}
}

func (c *converter) walk(n *html.Node) {
	switch n.Type {
	case html.CommentNode, html.DoctypeNode:
		return
	case html.TextNode:
		c.writeText(n.Data)
		return
	}

	if n.Type != html.ElementNode {
		c.walkChildren(n)
		return
	}

	if c.dropped(n) {
		return
	}

	switch n.Data {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(n.Data[1] - '0')
		c.blockGap()
		c.writeString(strings.Repeat("#", level) + " ")
		c.walkChildren(n)
		c.blockGap()

	case "p", "div", "section", "article", "main", "figure", "figcaption":
		c.blockGap()
		c.walkChildren(n)
		c.blockGap()

	case "br":
		c.writeString("  \n")

	case "hr":
		c.blockGap()
		c.writeString("---")
		c.blockGap()

	case "strong", "b":
		c.wrapInline(n, "**", "**")

	case "em", "i":
		c.wrapInline(n, "*", "*")

	case "s", "del":
		c.wrapInline(n, "~~", "~~")

	case "code":
		if isInsidePre(n) {
			c.walkChildren(n)
		} else {
			c.wrapInline(n, "`", "`")
		}

	case "pre":
		c.blockGap()
		c.writeString("```\n")
		c.writeString(c.textContent(n))
		c.writeString("\n```")
		c.blockGap()

	case "a":
		href, _ := attr(n, "href")
		href = strings.TrimSpace(href)
		text := strings.TrimSpace(collapseSpace(c.textContent(n)))
		if text == "" {
			return
		}
		if href == "" {
			c.writeString(text)
		} else {
			c.writeString("[" + text + "](" + href + ")")
		}

	case "img":
		alt, _ := attr(n, "alt")
		src, _ := attr(n, "src")
		alt = strings.TrimSpace(collapseSpace(alt))
		src = strings.TrimSpace(src)
		if src == "" {
			return
		}
		c.blockGap()
		c.writeString("![" + alt + "](" + src + ")")
		c.blockGap()

	case "ul", "ol":
		c.blockGap()
		c.listDepth++
		c.orderedIdx = append(c.orderedIdx, 0)
		for li := n.FirstChild; li != nil; li = li.NextSibling {
			if li.Type != html.ElementNode || li.Data != "li" {
				continue
			}
			c.orderedIdx[len(c.orderedIdx)-1]++
			indent := strings.Repeat("  ", c.listDepth-1)
			var marker string
			if n.Data == "ol" {
				marker = strconv.Itoa(c.orderedIdx[len(c.orderedIdx)-1]) + ". "
			} else {
				marker = "- "
			}
			c.writeString(indent + marker)
			c.atLineStart = false
			c.walkChildren(li)
			s := strings.TrimRight(c.buf.String(), " \t")
			c.buf.Reset()
			c.buf.WriteString(s)
			if !strings.HasSuffix(s, "\n") {
				c.writeString("\n")
			}
		}
		c.orderedIdx = c.orderedIdx[:len(c.orderedIdx)-1]
		c.listDepth--
		c.blockGap()

	case "blockquote":
		c.blockGap()
		inner := &converter{filters: c.filters}
		inner.walkChildren(n)
		lines := strings.Split(strings.TrimSpace(collapseBlankLines(inner.buf.String())), "\n")
		for _, l := range lines {
			c.writeString("> " + l + "\n")
		}
		c.blockGap()

	case "table":
		c.blockGap()
		c.writeTable(n)
		c.blockGap()

	default:
		c.walkChildren(n)
	}
}

func (c *converter) wrapInline(n *html.Node, open, close string) {
	text := strings.TrimSpace(collapseSpace(c.textContent(n)))
	if text == "" {
		return
	}
	c.writeString(open + text + close)
}

func (c *converter) writeText(s string) {
	collapsed := collapseSpace(s)
	if collapsed == "" {
		return
	}
	if c.atLineStart {
		collapsed = strings.TrimLeft(collapsed, " ")
		if collapsed == "" {
			return
		}
	}
	c.writeString(collapsed)
}

func (c *converter) writeTable(n *html.Node) {
	var rows [][]string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		for ch := node.FirstChild; ch != nil; ch = ch.NextSibling {
			if ch.Type != html.ElementNode {
				continue
			}
			switch ch.Data {
			case "thead", "tbody", "tfoot":
				walk(ch)
			case "tr":
				var row []string
				for cell := ch.FirstChild; cell != nil; cell = cell.NextSibling {
					if cell.Type == html.ElementNode && (cell.Data == "td" || cell.Data == "th") {
						row = append(row, strings.TrimSpace(collapseSpace(c.textContent(cell))))
					}
				}
				rows = append(rows, row)
			}
		}
	}
	walk(n)

	if len(rows) == 0 {
		return
	}
	cols := 0
	for _, r := range rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	pad := func(r []string) []string {
		for len(r) < cols {
			r = append(r, "")
		}
		return r
	}

	c.writeString("| " + strings.Join(pad(rows[0]), " | ") + " |\n")
	sep := make([]string, cols)
	for i := range sep {
		sep[i] = "---"
	}
	c.writeString("| " + strings.Join(sep, " | ") + " |\n")
	for _, r := range rows[1:] {
		c.writeString("| " + strings.Join(pad(r), " | ") + " |\n")
	}
}

func attr(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

// textContent flattens n's descendant text, skipping noise elements and any
// element matched by c's filter selectors (e.g. icon-font glyphs).
func (c *converter) textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
			return
		}
		if node.Type == html.ElementNode && node.Data == "br" {
			b.WriteString("\n")
			return
		}
		if c.dropped(node) {
			return
		}
		for ch := node.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(n)
	return b.String()
}

func isInsidePre(n *html.Node) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && p.Data == "pre" {
			return true
		}
	}
	return false
}

func collapseSpace(s string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !prevSpace {
				b.WriteByte(' ')
			}
			prevSpace = true
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

func collapseBlankLines(s string) string {
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return s
}
