// Command mdify scans a directory of HTML/PHP files and writes clean
// Markdown equivalents to an output directory, mirroring the input layout.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mdify/internal/htmlmd"
	"mdify/internal/phpstrip"
	"mdify/internal/rewrite"
	"mdify/internal/walker"
)

// repeatableFlag collects every occurrence of a flag passed multiple times,
// e.g. -replace 'a=>b' -replace 'c=>d'.
type repeatableFlag []string

func (r *repeatableFlag) String() string { return strings.Join(*r, ",") }
func (r *repeatableFlag) Set(v string) error {
	*r = append(*r, v)
	return nil
}

func main() {
	var (
		outDir         string
		extList        string
		stripSelectors string
		noIconFilters  bool
		rulesFile      string
		verbose        bool
	)
	var replaceSpecs repeatableFlag

	flag.StringVar(&outDir, "o", "md-output", "output directory")
	flag.StringVar(&extList, "ext", "html,htm,php", "comma-separated list of file extensions to convert")
	flag.StringVar(&stripSelectors, "strip-selector", "", "comma-separated selectors for elements to drop entirely, e.g. \"span.my-icon,i.custom-icon,.badge-*\"")
	flag.BoolVar(&noIconFilters, "no-icon-filters", false, "disable the built-in icon-font filters (Material Icons/Symbols, Font Awesome, Bootstrap Icons, Octicons)")
	flag.Var(&replaceSpecs, "replace", "regex find/replace applied to each file's Markdown output, as PATTERN=>REPLACEMENT (repeatable, applied in order given)")
	flag.StringVar(&rulesFile, "rules-file", "", "JSON file of [{\"pattern\":\"...\",\"replace\":\"...\"}, ...] rules, applied before -replace flags")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: mdify [flags] <input-dir>\n\nflags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	inputDir := flag.Arg(0)

	ext := map[string]bool{}
	for _, e := range strings.Split(extList, ",") {
		e = strings.TrimSpace(strings.ToLower(e))
		if e != "" {
			ext[e] = true
		}
	}

	convOpts := htmlmd.Options{
		ExtraSelectors:      htmlmd.ParseSelectors(stripSelectors),
		DisableDefaultIcons: noIconFilters,
	}

	var rules []rewrite.Rule
	if rulesFile != "" {
		fileRules, err := rewrite.LoadRulesFile(rulesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mdify: %v\n", err)
			os.Exit(1)
		}
		rules = append(rules, fileRules...)
	}
	if len(replaceSpecs) > 0 {
		flagRules, err := rewrite.ParseRules(replaceSpecs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mdify: %v\n", err)
			os.Exit(1)
		}
		rules = append(rules, flagRules...)
	}

	files, err := walker.Find(inputDir, ext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mdify: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "mdify: no matching files found under %s (extensions: %s)\n", inputDir, extList)
		os.Exit(1)
	}

	var failed, skipped int
	for _, f := range files {
		wrote, err := convertFile(f, outDir, convOpts, rules)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mdify: %s: %v\n", f.Src, err)
			failed++
			continue
		}
		if !wrote {
			skipped++
			if verbose {
				fmt.Printf("%s -> (empty, skipped)\n", f.Src)
			}
			continue
		}
		if verbose {
			fmt.Printf("%s -> %s\n", f.Src, filepath.Join(outDir, f.Out))
		}
	}

	fmt.Printf("mdify: converted %d/%d files into %s (%d skipped: empty)\n", len(files)-failed-skipped, len(files), outDir, skipped)
	if failed > 0 {
		os.Exit(1)
	}
}

// convertFile converts f, applies rules to the resulting Markdown, and
// writes it under outDir. It returns wrote=false without creating a file
// when the converted Markdown has no content.
func convertFile(f walker.File, outDir string, opts htmlmd.Options, rules []rewrite.Rule) (wrote bool, err error) {
	src, err := os.ReadFile(f.Src)
	if err != nil {
		return false, err
	}

	if strings.EqualFold(filepath.Ext(f.Src), ".php") {
		src = phpstrip.Strip(src)
	}

	md, err := htmlmd.ConvertWithOptions(src, opts)
	if err != nil {
		return false, err
	}

	if len(rules) > 0 {
		md = rewrite.Apply(rules, md)
	}

	if strings.TrimSpace(md) == "" {
		return false, nil
	}

	outPath := filepath.Join(outDir, f.Out)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		return false, err
	}
	return true, nil
}
