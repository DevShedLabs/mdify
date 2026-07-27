# mdify

> HTML/PHP → Markdown transpiler, optimized for CLI workflows

Scans a directory of website source (`.html`, `.htm`, `.php`), strips
non-content chrome (nav, header, footer, scripts, styles) and PHP code, and
writes clean Markdown files to an output directory that mirrors the input
layout — handy for feeding a site's content to an LLM.

## Install

```sh
go install github.com/DevShedLabs/mdify@latest

# or specific version
go install github.com/DevShedLabs/mdify@v0.0.4
```

## Updating
```sh 
mdify --version

mdify update 
```

## Usage

```sh
go build -o mdify .
./mdify -o md-output ./path/to/site
```

Flags:

- `-o` output directory (default `md-output`)
- `-ext` comma-separated extensions to convert (default `html,htm,php`)
- `-strip-selector` comma-separated selectors for elements to drop entirely,
  e.g. `"span.my-icon,i.custom-icon,.badge-*"` (trailing `*` matches a class
  prefix)
- `-no-icon-filters` disable the built-in icon-font filters
- `-replace` regex find/replace applied to each file's Markdown output, as
  `PATTERN=>REPLACEMENT` (repeatable, applied in the order given)
- `-rules-file` JSON file of `[{"pattern": "...", "replace": "..."}, ...]`
  rules, applied before `-replace` flags
- `-v` verbose per-file output
- `-version` print the installed version and check the Go module proxy for
  a newer release

Commands:

- `mdify update` reinstall the latest version (runs
  `go install github.com/DevShedLabs/mdify@latest`; requires a Go toolchain
  on `PATH`, same as the original install)

## Notes

- PHP code (`<?php ... ?>`, `<?= ... ?>`) is stripped; only the static HTML
  around it is converted. Dynamic content isn't rendered.
- `<script>`, `<style>`, `<nav>`, `<header>`, `<footer>`, `<aside>`, `<form>`,
  and similar chrome elements are dropped before conversion.
- Icon-font elements are dropped by default, since their ligature/glyph text
  (e.g. `arrow_back`) isn't real content: Material Icons/Symbols
  (`.material-icons`, `.material-symbols-outlined`, ...), Font Awesome
  (`i.fa`/`i.fas`/`i.far`/`i.fal`/`i.fab`/`i.fad`, any `.fa-*` class),
  Bootstrap Icons (`.bi-*`), `.glyphicon-*`, and `.octicon-*`. Use
  `-strip-selector` to add your own, or `-no-icon-filters` to disable the
  defaults.
- PHP variables (e.g. a `$site_url` base-path prefix) are gone once PHP is
  stripped, which can leave broken-looking links behind, like
  `[Read →](docs/editor/)` where the site actually needed `editor/`, or
  `[All docs](/docs/)` where the output tree needs `../` instead. This is
  project-specific and can't be inferred, so fix it with `-replace`/
  `-rules-file` regex rules run over the final Markdown, e.g.:

  ```sh
  ./mdify -o md-output \
    -replace 'docs/editor/=>editor/' \
    -replace '\(/docs/\)=>(../)' \
    ./path/to/site
  ```

  Patterns are Go regexp syntax; replacements may use `$1`/`${name}` capture
  references. Rules run in order, so later rules see earlier rules' output.
  Note Go regexp `^`/`$` anchor to the whole string by default, not each
  line — use `(?m)` in the pattern if you need per-line anchoring.

Built on `golang.org/x/net/html`.
