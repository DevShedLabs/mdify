package htmlmd

import (
	"strings"
	"testing"
)

func TestDefaultIconFiltersDropMaterialIcons(t *testing.T) {
	out := mustConvert(t, `<a href="/docs/"><span class="material-icons">arrow_back</span> All docs</a>`)
	if strings.Contains(out, "arrow_back") {
		t.Errorf("expected material icon ligature stripped, got: %q", out)
	}
	if !strings.Contains(out, "[All docs](/docs/)") {
		t.Errorf("expected link text preserved without icon, got: %q", out)
	}
}

func TestDefaultIconFiltersDropFontAwesome(t *testing.T) {
	out := mustConvert(t, `<p><i class="fa fa-home"></i> Home</p>`)
	if strings.Contains(out, "fa-home") {
		t.Errorf("got: %q", out)
	}
	if !strings.Contains(out, "Home") {
		t.Errorf("expected surrounding text preserved, got: %q", out)
	}
}

func TestExtraSelectorDropsCustomIcon(t *testing.T) {
	src := []byte(`<p><span class="my-icon">X</span> Keep this</p>`)
	out, err := ConvertWithOptions(src, Options{ExtraSelectors: ParseSelectors("span.my-icon")})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "X") {
		t.Errorf("expected custom icon span dropped, got: %q", out)
	}
	if !strings.Contains(out, "Keep this") {
		t.Errorf("expected surrounding text kept, got: %q", out)
	}
}

func TestDisableDefaultIcons(t *testing.T) {
	src := []byte(`<span class="material-icons">arrow_back</span>`)
	out, err := ConvertWithOptions(src, Options{DisableDefaultIcons: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "arrow_back") {
		t.Errorf("expected icon text kept when defaults disabled, got: %q", out)
	}
}

func TestSelectorClassPrefix(t *testing.T) {
	sel := ParseSelector(".fa-*")
	if !sel.ClassPrefix || sel.Class != "fa-" || sel.Tag != "" {
		t.Errorf("got %+v", sel)
	}
}

func TestSelectorTagAndClass(t *testing.T) {
	sel := ParseSelector("i.fa")
	if sel.Tag != "i" || sel.Class != "fa" || sel.ClassPrefix {
		t.Errorf("got %+v", sel)
	}
}

func TestParseSelectorsList(t *testing.T) {
	sels := ParseSelectors("span.a, i.b , .c-*")
	if len(sels) != 3 {
		t.Fatalf("got %d selectors: %+v", len(sels), sels)
	}
}
