package selfupdate

import "testing"

func TestEscapePath(t *testing.T) {
	cases := map[string]string{
		"github.com/DevShedLabs/mdify": "github.com/!dev!shed!labs/mdify",
		"github.com/foo/bar":           "github.com/foo/bar",
		"":                             "",
	}
	for in, want := range cases {
		if got := escapePath(in); got != want {
			t.Errorf("escapePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCurrentVersionDuringTest(t *testing.T) {
	// `go test` builds the main module without a version tag, so this
	// should fall back to "dev" rather than panic or return "(devel)".
	if v := CurrentVersion(); v == "(devel)" || v == "" {
		t.Errorf("expected dev fallback, got %q", v)
	}
}
