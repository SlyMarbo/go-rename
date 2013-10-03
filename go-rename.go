package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Match struct {
	Path  string
	Match string
}

type Matcher interface {
	Match(path string) *Match
	Regex() *regexp.Regexp
	fmt.Stringer
}

func Scan(dir string, matcher Matcher, recurse bool) (matches []*Match, err error) {
	dirs := []string{dir}

	for len(dirs) > 0 {
		dir := dirs[0]

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			path := filepath.Join(dir, file.Name())
			match := matcher.Match(path)
			if match != nil {
				matches = append(matches, match)
			}
			if recurse && file.IsDir() {
				dirs = append(dirs, path)
			}
		}

		dirs = dirs[1:]
	}

	return matches, nil
}

func ReplaceString(match *Match, replacement string) error {
	after := strings.Replace(match.Path, match.Match, replacement, 1)

	if *verbose && !*testing {
		fmt.Printf("Replacing \"%s\" with \"%s\".\n", match.Path, after)
	}

	if !*testing {
		return os.Rename(match.Path, after)
	} else {
		fmt.Printf("Would have replaced \"%s\" with \"%s\".\n", match.Path, after)
	}

	return nil
}

func ReplaceRegex(match *Match, replacement string, regex *regexp.Regexp) error {
	if regex == nil {
		return errors.New("-toRegex must be used with -fromRegex")
	}

	matches := regex.FindStringSubmatch(match.Path)
	if matches == nil || len(matches) == 0 {
		return errors.New("-toRegex does not match structure of -fromRegex")
	}

	defer func() {
		if err := recover(); err != nil {
			panic("-toRegex does not match structure of -fromRegex")
		}
	}()

	parts := make([]interface{}, len(matches)-1)
	for i, match := range matches[1:] {
		parts[i] = match
	}

	replacement = fmt.Sprintf(replacement, parts...)

	after := strings.Replace(match.Path, match.Match, replacement, 1)

	if *verbose && !*testing {
		fmt.Printf("Replacing \"%s\" with \"%s\".\n", match.Path, after)
	}

	if !*testing {
		return os.Rename(match.Path, after)
	} else {
		fmt.Printf("Would have replaced \"%s\" with \"%s\".\n", match.Path, after)
	}

	return nil
}

type StringMatcher struct {
	selector string
	ext      string
}

func NewStringMatcher(selector, ext string) Matcher {
	return &StringMatcher{selector, ext}
}

func (m *StringMatcher) Match(path string) *Match {
	if (m.ext == "" || m.ext == filepath.Ext(path)) && strings.Contains(path, m.selector) {
		return &Match{path, m.selector}
	}
	return nil
}

func (_ *StringMatcher) Regex() *regexp.Regexp {
	return nil
}

func (m *StringMatcher) String() string {
	if m.ext == "" {
		return fmt.Sprintf("Matching string \"%s\".", m.selector)
	}
	return fmt.Sprintf("Matching string \"%s\" with extension \"%s\".", m.selector, m.ext)
}

type RegexMatcher struct {
	regex *regexp.Regexp
	ext   string
}

func NewRegexMatcher(selector, ext string) Matcher {
	selector = strings.Replace(selector, "%", "\\", -1)
	selector = strings.Replace(selector, "\\\\", "%", -1)
	return &RegexMatcher{
		regex: regexp.MustCompile(selector),
		ext:   ext,
	}
}

func (m *RegexMatcher) Match(path string) *Match {
	if m.ext == "" || m.ext == filepath.Ext(path) {
		matches := m.regex.FindStringSubmatch(path)
		if matches == nil {
			return nil
		}

		return &Match{path, matches[0]}
	}

	return nil
}

func (m *RegexMatcher) Regex() *regexp.Regexp {
	return m.regex
}

func (m *RegexMatcher) String() string {
	if m.ext == "" {
		return fmt.Sprintf("Matching regex /%s/.", m.regex.String())
	}
	return fmt.Sprintf("Matching regex /%s/ with extension \"%s\".", m.regex.String(), m.ext)
}

var start = flag.String("start", ".", "starting directory")
var ext = flag.String("ext", "", "file extension (with dot)")
var from = flag.String("from", "", "initial form (simple string)")
var fromR = flag.String("fromRegex", "", "initial form (regex)")
var to = flag.String("to", "", "final form (simple string)")
var toR = flag.String("toRegex", "", "final form (regex)")
var recurse = flag.Bool("r", false, "whether to recurse into child folders")
var verbose = flag.Bool("v", false, "whether to enable verbose feedback")
var testing = flag.Bool("test", false, "prints renaming operations without enacting them")

func main() {
	flag.Parse()

	var matcher Matcher
	if *from != "" {
		matcher = NewStringMatcher(*from, *ext)
	} else if *fromR != "" {
		matcher = NewRegexMatcher(*fromR, *ext)
	} else {
		fmt.Fprintln(os.Stderr, "Error: No selector provided. Use -from or -fromRegex")
		return
	}

	if *verbose {
		fmt.Println(matcher)
	}

	matches, err := Scan(*start, matcher, *recurse)
	if err != nil {
		fmt.Printf("Error: %q.\n", err)
		return
	}

	if matches == nil || len(matches) == 0 {
		fmt.Println("No matches found.")
		return
	}

	for _, match := range matches {
		if *verbose {
			fmt.Printf("Found match \"%s\".\n", match.Path)
		}

		if *to != "" {
			err = ReplaceString(match, *to)
			if err != nil {
				panic(err)
			}
		} else if *toR != "" {
			err = ReplaceRegex(match, *toR, matcher.Regex())
			if err != nil {
				panic(err)
			}
		} else {
			err = ReplaceString(match, "")
			if err != nil {
				panic(err)
			}
		}
	}
}
