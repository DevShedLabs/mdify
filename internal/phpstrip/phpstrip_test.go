package phpstrip

import (
	"strings"
	"testing"
)

func TestStripBasic(t *testing.T) {
	src := `<html><?php require 'x.php'; ?><body><p>Hi <?= $name ?></p></body></html>`
	out := string(Strip([]byte(src)))
	if strings.Contains(out, "php") || strings.Contains(out, "$name") {
		t.Errorf("php not stripped: %q", out)
	}
	if !strings.Contains(out, "<html>") || !strings.Contains(out, "<p>Hi") {
		t.Errorf("static html lost: %q", out)
	}
}

func TestUnterminatedBlock(t *testing.T) {
	src := `<p>before</p><?php echo "no close"`
	out := string(Strip([]byte(src)))
	if !strings.Contains(out, "<p>before</p>") {
		t.Errorf("expected static prefix kept, got %q", out)
	}
	if strings.Contains(out, "no close") {
		t.Errorf("expected unterminated php dropped, got %q", out)
	}
}

func TestShortTag(t *testing.T) {
	src := `A<? echo 1; ?>B`
	out := string(Strip([]byte(src)))
	if out != "A B" {
		t.Errorf("got %q", out)
	}
}
