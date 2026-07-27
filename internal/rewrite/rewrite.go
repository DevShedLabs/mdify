// Package rewrite applies user-supplied regex find/replace rules to
// converted Markdown, as a post-processing pass. This is how project-specific
// breakage — e.g. leftover PHP base-path variables collapsing a link like
// [Read →](docs/editor/) into a path that no longer resolves from the
// output's location — gets patched, since the correct fix is different for
// every project and can't be inferred generically.
package rewrite

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Rule is a single regex find/replace. Replace uses Go's regexp.Expand
// syntax ($1, ${name}) to reference capture groups from Pattern.
type Rule struct {
	Pattern *regexp.Regexp
	Replace string
}

// ParseRule parses "PATTERN=>REPLACEMENT", where PATTERN is a Go regexp and
// REPLACEMENT may reference its capture groups ($1, ${name}).
func ParseRule(s string) (Rule, error) {
	idx := strings.Index(s, "=>")
	if idx == -1 {
		return Rule{}, fmt.Errorf("invalid rule %q: expected PATTERN=>REPLACEMENT", s)
	}
	patStr, repl := s[:idx], s[idx+2:]

	re, err := regexp.Compile(patStr)
	if err != nil {
		return Rule{}, fmt.Errorf("invalid pattern %q: %w", patStr, err)
	}
	return Rule{Pattern: re, Replace: repl}, nil
}

// ParseRules parses multiple "PATTERN=>REPLACEMENT" strings, in order.
func ParseRules(specs []string) ([]Rule, error) {
	rules := make([]Rule, 0, len(specs))
	for _, s := range specs {
		r, err := ParseRule(s)
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

type fileRule struct {
	Pattern string `json:"pattern"`
	Replace string `json:"replace"`
}

// LoadRulesFile reads a JSON array of {"pattern": "...", "replace": "..."}
// objects, applied top to bottom.
func LoadRulesFile(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw []fileRule
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	rules := make([]Rule, 0, len(raw))
	for _, fr := range raw {
		re, err := regexp.Compile(fr.Pattern)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid pattern %q: %w", path, fr.Pattern, err)
		}
		rules = append(rules, Rule{Pattern: re, Replace: fr.Replace})
	}
	return rules, nil
}

// Apply runs each rule's regex replace over text, in order.
func Apply(rules []Rule, text string) string {
	for _, r := range rules {
		text = r.Pattern.ReplaceAllString(text, r.Replace)
	}
	return text
}
