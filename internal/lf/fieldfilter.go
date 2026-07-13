package lf

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// compiledPatternCache memoizes compiled exclude-fields regexes across Run calls.
// Keyed by the raw pattern string; values are *regexp.Regexp.
// Invalid patterns are rejected earlier by Config.Validate (and by flag parsing),
// so failures here are silently skipped rather than re-reported per field.
var compiledPatternCache sync.Map

// compileFieldPatterns returns compiled regexes for the given patterns,
// skipping any that fail to compile.
func compileFieldPatterns(patterns []string) []*regexp.Regexp {
	if len(patterns) == 0 {
		return nil
	}
	res := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if p == "" {
			continue
		}
		if cached, ok := compiledPatternCache.Load(p); ok {
			if re, isRe := cached.(*regexp.Regexp); isRe {
				res = append(res, re)
			}
			continue
		}
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		compiledPatternCache.Store(p, re)
		res = append(res, re)
	}
	return res
}

// isFieldExcluded reports whether a field should be excluded from validation
// based on exclude-fields regex patterns. Patterns are matched (unanchored,
// standard regexp semantics) against both the leaf field name (e.g. "CreatedAt")
// and the full nested path (e.g. "User.Role.CreatedAt").
func isFieldExcluded(leafName, fullPath string, patterns []*regexp.Regexp) bool {
	for _, re := range patterns {
		if re.MatchString(leafName) || re.MatchString(fullPath) {
			return true
		}
	}
	return false
}

// isFieldTagIgnored reports whether a struct field's tag matches any ignore-tags entry.
//
// Each entry is either:
//   - a bare tag key (e.g. "lostfield"): the field is ignored if the key is present,
//     regardless of its value;
//   - key:"value" form (e.g. `lostfield:"ignore"` or `json:"-"`): the field is ignored
//     only when the tag value matches exactly. The value may be quoted or bare.
func isFieldTagIgnored(structTag string, ignoreTags []string) bool {
	if structTag == "" || len(ignoreTags) == 0 {
		return false
	}
	tag := reflect.StructTag(structTag)

	for _, entry := range ignoreTags {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		key, rawVal, hasVal := strings.Cut(entry, ":")
		val, ok := tag.Lookup(key)
		if !ok {
			continue
		}
		if !hasVal {
			// Bare key: presence is enough.
			return true
		}
		want := rawVal
		if unquoted, err := strconv.Unquote(rawVal); err == nil {
			want = unquoted
		}
		if val == want {
			return true
		}
	}
	return false
}

// typeNameSimilarity computes the Sørensen–Dice bigram coefficient between two
// type names (case-insensitive). Returns a value in [0.0, 1.0], where 1.0 means
// identical names. Names shorter than two characters are compared by equality.
func typeNameSimilarity(a, b string) float64 {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	if a == b {
		return 1.0
	}
	if len(a) < 2 || len(b) < 2 {
		return 0.0
	}

	bigrams := func(s string) map[string]int {
		m := make(map[string]int, len(s)-1)
		runes := []rune(s)
		for i := 0; i+1 < len(runes); i++ {
			m[string(runes[i:i+2])]++
		}
		return m
	}

	aBigrams := bigrams(a)
	bBigrams := bigrams(b)

	var totalA, totalB, common int
	for _, n := range aBigrams {
		totalA += n
	}
	for bg, n := range bBigrams {
		totalB += n
		if an, ok := aBigrams[bg]; ok {
			common += min(an, n)
		}
	}

	return 2.0 * float64(common) / float64(totalA+totalB)
}
