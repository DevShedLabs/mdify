package htmlmd

import (
	"strings"
	"testing"
)

func mustConvert(t *testing.T, src string) string {
	t.Helper()
	out, err := Convert([]byte(src))
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	return out
}

func TestStripsChrome(t *testing.T) {
	out := mustConvert(t, `<html><body>
		<nav>Nav</nav><header>Header</header>
		<main><p>Content</p></main>
		<footer>Footer</footer>
		<script>evil()</script><style>.x{}</style>
	</body></html>`)

	for _, noiseWord := range []string{"Nav", "Header", "Footer", "evil()"} {
		if strings.Contains(out, noiseWord) {
			t.Errorf("expected %q to be stripped, got:\n%s", noiseWord, out)
		}
	}
	if !strings.Contains(out, "Content") {
		t.Errorf("expected content preserved, got:\n%s", out)
	}
}

func TestHeadingsAndInline(t *testing.T) {
	out := mustConvert(t, `<h1>Title</h1><p>A <strong>bold</strong> and <em>italic</em> word, plus <a href="/x">a link</a>.</p>`)
	want := "# Title\n\nA **bold** and *italic* word, plus [a link](/x).\n"
	if out != want {
		t.Errorf("got:\n%q\nwant:\n%q", out, want)
	}
}

func TestLists(t *testing.T) {
	out := mustConvert(t, `<ul><li>One</li><li>Two</li></ul><ol><li>First</li><li>Second</li></ol>`)
	if !strings.Contains(out, "- One\n- Two") {
		t.Errorf("unordered list wrong:\n%s", out)
	}
	if !strings.Contains(out, "1. First\n2. Second") {
		t.Errorf("ordered list wrong:\n%s", out)
	}
}

func TestCodeBlock(t *testing.T) {
	out := mustConvert(t, "<pre><code>line1\nline2</code></pre>")
	want := "```\nline1\nline2\n```\n"
	if out != want {
		t.Errorf("got:\n%q\nwant:\n%q", out, want)
	}
}

func TestTable(t *testing.T) {
	out := mustConvert(t, `<table><tr><th>A</th><th>B</th></tr><tr><td>1</td><td>2</td></tr></table>`)
	want := "| A | B |\n| --- | --- |\n| 1 | 2 |\n"
	if out != want {
		t.Errorf("got:\n%q\nwant:\n%q", out, want)
	}
}

func TestImage(t *testing.T) {
	out := mustConvert(t, `<img src="/a.png" alt="Alt text">`)
	if !strings.Contains(out, "![Alt text](/a.png)") {
		t.Errorf("got:\n%s", out)
	}
}

func TestNoBlankLineExplosion(t *testing.T) {
	out := mustConvert(t, `<p>A</p>



	<p>B</p>`)
	if strings.Contains(out, "\n\n\n") {
		t.Errorf("expected no triple newlines, got:\n%q", out)
	}
}
