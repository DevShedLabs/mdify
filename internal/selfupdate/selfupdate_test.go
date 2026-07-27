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

func TestVersionCheckMessage(t *testing.T) {
	cases := []struct {
		name, cur, latest, want string
	}{
		{"equal", "v0.0.3", "v0.0.3", "up to date"},
		{"update available", "v0.0.2", "v0.0.3", "update available: v0.0.2 -> v0.0.3 (run `mdify update`)"},
		{"proxy lagging behind installed version", "v0.0.3", "v0.0.2", "up to date"},
		{"dev build", "dev", "v0.0.3", "latest published version: v0.0.3"},
		{"dirty build metadata treated as equal", "v0.0.2+dirty", "v0.0.2", "up to date"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := versionCheckMessage(c.cur, c.latest); got != c.want {
				t.Errorf("versionCheckMessage(%q, %q) = %q, want %q", c.cur, c.latest, got, c.want)
			}
		})
	}
}

func TestCurrentVersionDuringTest(t *testing.T) {
	// `go test` builds the main module without a version tag, so this
	// should fall back to "dev" rather than panic or return "(devel)".
	if v := CurrentVersion(); v == "(devel)" || v == "" {
		t.Errorf("expected dev fallback, got %q", v)
	}
}
