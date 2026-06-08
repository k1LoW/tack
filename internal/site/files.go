package site

import (
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// Entry is one file to upload as part of a deployment.
type Entry struct {
	AbsPath     string
	RelPath     string
	ContentType string
}

var ignoredBasenames = map[string]struct{}{
	".ds_store":   {},
	"thumbs.db":   {},
	"desktop.ini": {},
}

func shouldIgnore(name string) bool {
	_, ok := ignoredBasenames[strings.ToLower(filepath.Base(name))]
	return ok
}

// Collect resolves the given path into a list of upload entries. If the path
// is a directory it is walked recursively; if it is a single file, just that
// file is returned.
func Collect(path string) ([]Entry, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("path does not exist: %s", path)
		}
		return nil, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsDir() {
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("path is not a regular file or directory: %s", path)
		}
		if shouldIgnore(abs) {
			return nil, fmt.Errorf("path is an ignored file: %s", path)
		}
		return []Entry{{
			AbsPath:     abs,
			RelPath:     filepath.Base(abs),
			ContentType: contentTypeFor(abs),
		}}, nil
	}

	var entries []Entry
	walkErr := filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if shouldIgnore(p) {
			return nil
		}
		rel, err := filepath.Rel(abs, p)
		if err != nil {
			return err
		}
		entries = append(entries, Entry{
			AbsPath:     p,
			RelPath:     filepath.ToSlash(rel),
			ContentType: contentTypeFor(p),
		})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no deployable files found under %s", path)
	}
	return entries, nil
}

func contentTypeFor(p string) string {
	if ct := mime.TypeByExtension(filepath.Ext(p)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}
