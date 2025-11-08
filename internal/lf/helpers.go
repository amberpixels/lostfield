package lf

import (
	"path/filepath"
	"strings"
)

// MatchesAnyPattern checks if a name matches any of the glob patterns.
// For patterns containing '/', matches against the full path.
// For patterns without '/', matches against only the basename.
// Patterns like */vendor/* are matched using substring matching.
// Returns true if the name matches any pattern, false otherwise.
func MatchesAnyPattern(name string, patterns []string) bool {
	// Empty name never matches
	if name == "" {
		return false
	}

	for _, pattern := range patterns {
		var matched bool
		var err error

		// Special case: patterns like */vendor/* should match paths containing the substring
		// Since filepath.Match doesn't support multi-level globs well, use string matching
		if strings.HasPrefix(pattern, "*/") && strings.HasSuffix(pattern, "/*") {
			substring := pattern[1 : len(pattern)-1] // Remove leading */ and trailing /*
			if strings.Contains(filepath.ToSlash(name), substring) {
				return true
			}
			continue
		}

		// If pattern contains '/', match against full path
		// Otherwise, match against just the basename
		if strings.Contains(pattern, "/") {
			matched, err = filepath.Match(pattern, name)
		} else {
			matched, err = filepath.Match(pattern, filepath.Base(name))
		}

		if err == nil && matched {
			return true
		}
	}

	return false
}
