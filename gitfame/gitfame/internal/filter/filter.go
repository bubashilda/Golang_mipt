package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

import (
	_ "embed"
	"gitlab.com/slon/shad-go/gitfame/configs"
)

type languageExt struct {
	Name       string
	Type       string
	Extensions []string
}

type FileFilter struct {
	allowedExts   []string
	excludeGlobs  []string
	restrictGlobs []string
}

func New(exts, langs, excludeGlobs, restrictGlobs []string) (*FileFilter, error) {
	var langExts []languageExt
	err := json.Unmarshal(configs.LanguageExtensionsJSON, &langExts)

	if err != nil {
		return nil, fmt.Errorf("failed to parse language extensions: %v", err)
	}

	matchedExts := make([]string, 0, len(exts))
	matchedExts = append(matchedExts, exts...)

	for _, lang := range langExts {
		if slices.Contains(langs, strings.ToLower(lang.Name)) {
			matchedExts = append(matchedExts, lang.Extensions...)
		}
	}

	for _, pattern := range excludeGlobs {
		if _, err := filepath.Match(pattern, ""); errors.Is(err, filepath.ErrBadPattern) {
			return nil, fmt.Errorf("invalid glob pattern %s: %v", pattern, err)
		}
	}

	for _, pattern := range restrictGlobs {
		if _, err := filepath.Match(pattern, ""); errors.Is(err, filepath.ErrBadPattern) {
			return nil, fmt.Errorf("invalid glob pattern %s: %v", pattern, err)
		}
	}

	return &FileFilter{
		excludeGlobs:  excludeGlobs,
		restrictGlobs: restrictGlobs,
		allowedExts:   matchedExts,
	}, nil
}

func (f *FileFilter) Match(path string) bool {
	for i, ext := range f.allowedExts {
		if strings.HasSuffix(path, ext) {
			break
		}
		if i == len(f.allowedExts)-1 {
			return false
		}
	}
	for _, pattern := range f.excludeGlobs {
		if match, _ := filepath.Match(pattern, path); match {
			return false
		}
	}
	for i, pattern := range f.restrictGlobs {
		if match, _ := filepath.Match(pattern, path); match {
			break
		} else if i == len(f.restrictGlobs)-1 {
			return false
		}
	}
	return true
}

func (f *FileFilter) Filter(paths []string) []string {
	matched := make([]string, 0)
	for _, path := range paths {
		if f.Match(path) {
			matched = append(matched, path)
		}
	}
	return matched
}
