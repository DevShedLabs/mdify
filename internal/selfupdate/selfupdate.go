// Package selfupdate implements mdify's -version update check and update
// command: it reads the version Go stamped into the binary at install time,
// queries the Go module proxy for the latest published version, and (for
// Update) shells out to `go install <module>@latest`.
package selfupdate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"time"

	"golang.org/x/mod/semver"
)

// CurrentVersion returns the module version Go recorded when this binary
// was built via `go install <module>@<version>`, or "dev" for a local
// build (go run/go build within the module) where no such version exists.
func CurrentVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok || bi.Main.Version == "" || bi.Main.Version == "(devel)" {
		return "dev"
	}
	return bi.Main.Version
}

// LatestVersion queries the Go module proxy for modulePath's latest
// published version (the same source `go install ...@latest` consults).
func LatestVersion(modulePath string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := "https://proxy.golang.org/" + escapePath(modulePath) + "/@latest"

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("module proxy returned %s", resp.Status)
	}

	var info struct{ Version string }
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	return info.Version, nil
}

// escapePath applies Go's module-path escaping — each uppercase letter
// becomes '!' followed by its lowercase form — as required by the module
// proxy protocol (e.g. "DevShedLabs" -> "!dev!shed!labs").
func escapePath(path string) string {
	b := make([]byte, 0, len(path)+4)
	for i := 0; i < len(path); i++ {
		c := path[i]
		if c >= 'A' && c <= 'Z' {
			b = append(b, '!', c-'A'+'a')
		} else {
			b = append(b, c)
		}
	}
	return string(b)
}

// PrintVersionCheck prints the running binary's version and, best-effort,
// reports whether a newer version has been published.
func PrintVersionCheck(modulePath string) {
	cur := CurrentVersion()
	fmt.Printf("mdify %s\n", cur)

	latest, err := LatestVersion(modulePath)
	if err != nil {
		fmt.Printf("(could not check for updates: %v)\n", err)
		return
	}

	fmt.Println(versionCheckMessage(cur, latest))
}

// versionCheckMessage compares cur (the running binary's version) against
// latest (what the module proxy currently reports), using real semver
// ordering rather than string inequality — so a proxy that hasn't yet
// indexed a just-published tag (briefly reporting an older "latest" than
// what's already installed) can never be misreported as an available
// "update" to an older version.
func versionCheckMessage(cur, latest string) string {
	if cur == "dev" {
		return fmt.Sprintf("latest published version: %s", latest)
	}
	if !semver.IsValid(cur) || !semver.IsValid(latest) {
		if cur == latest {
			return "up to date"
		}
		return fmt.Sprintf("current: %s, latest published: %s", cur, latest)
	}
	if semver.Compare(cur, latest) < 0 {
		return fmt.Sprintf("update available: %s -> %s (run `mdify update`)", cur, latest)
	}
	return "up to date"
}

// Update shells out to `go install <modulePath>@latest`, requiring a Go
// toolchain on PATH — the same mechanism originally used to install mdify.
func Update(modulePath string) error {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go toolchain not found in PATH: %w", err)
	}

	fmt.Printf("Running: go install %s@latest\n", modulePath)
	cmd := exec.Command(goBin, "install", modulePath+"@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
