// Package phpstrip removes PHP code blocks from source text, leaving the
// surrounding static HTML intact so it can be parsed as plain HTML.
package phpstrip

import "strings"

// openTags are the PHP opening tag variants we recognize, checked longest
// first so "<?php" isn't mis-matched by the shorter "<?" style tags.
var openTags = []string{"<?php", "<?=", "<?"}

const closeTag = "?>"

// Strip removes every <?php ... ?>, <?= ... ?> and short-tag <? ... ?> block
// from src, replacing each with a single space so surrounding text doesn't
// get glued together. An unterminated PHP block (no closing "?>") is cut off
// at end of input.
func Strip(src []byte) []byte {
	var out strings.Builder
	s := string(src)

	for {
		start, tagLen := findOpenTag(s)
		if start == -1 {
			out.WriteString(s)
			break
		}
		out.WriteString(s[:start])
		out.WriteByte(' ')

		rest := s[start+tagLen:]
		end := strings.Index(rest, closeTag)
		if end == -1 {
			// Unterminated PHP block: drop the remainder of the file.
			break
		}
		s = rest[end+len(closeTag):]
	}

	return []byte(out.String())
}

// findOpenTag returns the byte offset of the earliest PHP opening tag in s
// and the length of that tag, or (-1, 0) if none is present.
func findOpenTag(s string) (int, int) {
	bestIdx := -1
	bestLen := 0
	for _, tag := range openTags {
		if idx := strings.Index(s, tag); idx != -1 {
			if bestIdx == -1 || idx < bestIdx || (idx == bestIdx && len(tag) > bestLen) {
				bestIdx = idx
				bestLen = len(tag)
			}
		}
	}
	return bestIdx, bestLen
}
