package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Ext         string       `long:"ext" description:"Restrict matches to those with the given file extension."`
	From        func(string) `long:"from" description:"Initial form (string)."`
	FromRegex   func(string) `long:"from-regex" description:"Initial form (regex)."`
	To          func(string) `long:"to" description:"Replacement (string)."`
	ToRegex     func(string) `long:"to-regex" description:"Replacement (regex)."`
	NumberStart func(string) `long:"number-start" description:"Perform a maths op (eg -4) on the first number found."`
	NumberEnd   func(string) `long:"number-end" description:"Perform a maths op (eg -4) on the last number found."`

	Recurse bool `short:"R" description:"Recurse into child folders."`
	Verbose bool `short:"v" long:"verbose" description:"Enable verbose output."`
	Testing bool `short:"t" long:"test" description:"Print renaming ops without performing them."`
	ZeroPad int  `short:"z" long:"zero-pad" description:"Pad --number-* results with zeros to the given width."`
	Num     int  `short:"n" default:"1" description:"Number of times to perform each op per filename (-1 for unlimited)."`
}

type File struct {
	From string
	To   string
	done bool
}

func main() {
	var (
		fromString string
		fromRegex  *regexp.Regexp
		matcher    Matcher
		ext        Matcher
		replacer   Replacer
		dir        = "."
	)

	checkMatcher := func() {
		if matcher != nil {
			complain(ErrTooManyMatchers)
		}
	}
	checkReplacer := func() {
		if replacer != nil {
			complain(ErrTooManyReplacers)
		}
	}

	// Set up replacer and matcher functions.
	opts.From = func(s string) {
		checkMatcher()
		fromString = s
		matcher = StringMatcher(s)
		debug("Using string matcher.")
	}
	opts.FromRegex = func(s string) {
		checkMatcher()
		fromRegex = (*regexp.Regexp)(NewRegexMatcher(s))
		matcher = (*RegexMatcher)(fromRegex)
		debug("Using regex matcher.")
	}
	opts.NumberStart = func(s string) {
		checkMatcher()
		checkReplacer()
		fromRegex = numStartRegex
		matcher = NumberMatcher(true)
		replacer = NewNumberReplacer(s, true)
		debug("Using number matcher/replacer (start).")
	}
	opts.NumberEnd = func(s string) {
		checkMatcher()
		checkReplacer()
		fromRegex = numEndRegex
		matcher = NumberMatcher(false)
		replacer = NewNumberReplacer(s, false)
		debug("Using number matcher/replacer (end).")
	}
	opts.To = func(s string) {
		checkReplacer()
		replacer = StringReplacer(s)
		debug("Using string replacer.")
	}
	opts.ToRegex = func(s string) {
		checkReplacer()
		if fromRegex != nil {
			replacer = NewRegexReplacer(s)
			debug("Using regex-to-regex replacer.")
		} else {
			replacer = StringReplacer(s)
			debug("Using string-to-regex replacer.")
		}
	}

	parser := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	parser.Usage = "[dir]"
	args, err := parser.Parse()
	handle(err)

	// Starting dir
	if len(args) > 0 {
		if len(args) > 1 {
			parser.WriteHelp(os.Stderr)
			os.Exit(2)
		}
		dir = args[0]
	}

	// Extension
	if opts.Ext != "" {
		ext = ExtMatcher(opts.Ext)
		debug(fmt.Sprintf("Using extension matcher %q.", opts.Ext))
	}

	if replacer == nil {
		debug(`Warning: Neither --to nor --to-regex used. Using --to "".`)
		replacer = StringReplacer("")
	}
	if matcher == nil {
		parser.WriteHelp(os.Stderr)
		os.Exit(2)
	}

	// Collect matching files.
	matches := make([]*File, 0, 10)
	dirs := []string{dir}
	for len(dirs) > 0 {
		dir := dirs[0]

		files, err := ioutil.ReadDir(dir)
		handle(err)

		for _, file := range files {
			name := file.Name()
			path := filepath.Join(dir, name)
			if file.IsDir() {
				if opts.Recurse {
					dirs = append(dirs, path)
				}
				continue
			}

			if ext != nil && !ext.Matches(name) {
				continue
			}

			if matcher.Matches(name) {
				debug("Matching file: " + path)
				to := replacer.Replace(name, fromString, fromRegex)
				match := new(File)
				match.From = path
				match.To = filepath.Join(filepath.Dir(path), to)
				matches = append(matches, match)
			}
		}

		dirs = dirs[1:]
	}

	// Process matching files.
	others := make(map[string]*File, len(matches))
	for _, match := range matches {
		exists := false
		if _, err := os.Stat(match.To); err == nil {
			exists = true
		}

		for _, other := range matches {
			if other == match {
				continue
			}
			if other.To == match.To {
				complain("Error: Multiple files being renamed to " + match.To + ".")
			}
			if other.To == match.From && match.To == other.From {
				complain(fmt.Sprintf("Error: Swapping not yet supported (%q -> %q, %q -> %q).",
					match.From, match.To, other.From, other.To))
			}
			if exists && other.From == match.To {
				exists = false
			}
		}

		if exists {
			complain(fmt.Sprintf("Error: Cannot rename %q (%q already exists).\n", match.From, match.To))
		}

		others[match.From] = match
	}

	// Perform the actual renaming.
	for _, file := range matches {
		Rename(file, others)
	}
}

func Rename(f *File, others map[string]*File) {
	if f.done {
		return
	}

	if other, ok := others[f.To]; ok {
		Rename(other, others)
	}

	if !f.done {
		f.done = true
		if opts.Testing {
			fmt.Printf(" %s -> %s\n", f.From, f.To)
		} else {
			if opts.Verbose {
				fmt.Printf(" %s -> %s\n", f.From, f.To)
			}
			os.Rename(f.From, f.To)
		}
	}
}

func complain(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func debug(msg string) {
	if opts.Verbose {
		fmt.Println(msg)
	}
}

func handle(err error) {
	if err != nil {
		complain(err.Error())
	}
}

const ErrTooManyMatchers = "Error: Only one matcher can be used (--from, --from-regex, --number-start, --number-end)."
const ErrTooManyReplacers = "Error: Only one replacer can be used (--to, --to-regex, --number-start, --number-end)."
