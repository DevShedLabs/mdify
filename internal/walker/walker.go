// Package walker discovers input files under a directory tree and computes
// their mirrored output paths.
package walker

import (
	"os"
	"path/filepath"
	"strings"
)

// skipDirs holds directory names that are never worth descending into.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
}

// File pairs a discovered source file with its mirrored Markdown output path.
type File struct {
	Src string // absolute or relative path as passed in
	Out string // output path, relative to the output root
}

// Find walks root looking for files whose extension (without the dot,
// lowercased) is in ext. It returns each match paired with an output path
// that mirrors the input's relative location, with its extension replaced
// by .md.
func Find(root string, ext map[string]bool) ([]File, error) {
	var files []File

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		e := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
		if !ext[e] {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out := strings.TrimSuffix(rel, filepath.Ext(rel)) + ".md"

		files = append(files, File{Src: path, Out: out})
		return nil
	})

	return files, err
}
