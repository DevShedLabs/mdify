package rewrite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRuleAndApply(t *testing.T) {
	r, err := ParseRule(`\]\(docs/=>](editor-base/`)
	if err != nil {
		t.Fatal(err)
	}
	out := Apply([]Rule{r}, "[Read →](docs/editor/)")
	want := "[Read →](editor-base/editor/)"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestParseRuleInvalid(t *testing.T) {
	if _, err := ParseRule("no-delimiter-here"); err == nil {
		t.Fatal("expected error for missing =>")
	}
	if _, err := ParseRule("(unclosed=>x"); err == nil {
		t.Fatal("expected error for invalid regexp")
	}
}

func TestLoadRulesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	content := `[
		{"pattern": "^/docs/$", "replace": "../"},
		{"pattern": "\\(docs/", "replace": "(editor/"}
	]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadRulesFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 2 {
		t.Fatalf("got %d rules", len(rules))
	}

	out2 := Apply(rules, "[Read →](docs/editor/)")
	if out2 != "[Read →](editor/editor/)" {
		t.Errorf("got %q", out2)
	}
}

func TestApplyOrderMatters(t *testing.T) {
	rules, err := ParseRules([]string{"a=>b", "b=>c"})
	if err != nil {
		t.Fatal(err)
	}
	out := Apply(rules, "a")
	if out != "c" {
		t.Errorf("got %q, want c (rules should chain in order)", out)
	}
}
