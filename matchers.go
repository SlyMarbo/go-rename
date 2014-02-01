package main

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"
)

type Matcher interface {
	Matches(string) bool
}

type ExtMatcher string

func (e ExtMatcher) Matches(s string) bool {
	return strings.HasSuffix(s, string(e))
}

var numStartRegex = regexp.MustCompile(`\A\d+`)
var numEndRegex = regexp.MustCompile(`\d+\z`)

// If true, look from the start.
type NumberMatcher bool

func (n NumberMatcher) Matches(s string) bool {
	if n {
		return numStartRegex.MatchString(s)
	}
	ext := filepath.Ext(s)
	if ext != "" {
		i := strings.LastIndex(s, ext)
		s = s[:i]
	}
	return numEndRegex.MatchString(s)
}

type StringMatcher string

func (m StringMatcher) Matches(s string) bool {
	return strings.Contains(s, string(m))
}

type RegexMatcher regexp.Regexp

func NewRegexMatcher(s string) *RegexMatcher {
	buf := new(bytes.Buffer)
	for {
		i := strings.Index(s, "%")
		if i < 0 {
			break
		}

		buf.WriteString(s[:i])

		if len(s) == i+1 {
			buf.WriteString("%")
			break
		}

		switch s[i+1] {
		case 'd', 'D', 's', 'S', 'w', 'W':
			buf.Write([]byte{'\\', s[i+1]})
			s = s[i+2:]
			continue

		case '%':
			buf.WriteString("%")
			s = s[i+2:]
			continue

		default:
			complain("Error: Could not parse regex. " + s[i:i+2] + " is not a valid combination.")
		}
	}
	return (*RegexMatcher)(regexp.MustCompile(buf.String()))
}

func (r *RegexMatcher) Matches(s string) bool {
	return (*regexp.Regexp)(r).MatchString(s)
}
