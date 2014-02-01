# go-rename

A simple tool for clever mass-renaming files.

## Table of Contents

- [Installing go-rename](#installing-go-rename)
- [Using go-rename](#using-go-rename)
	- [Simple strings](#simple-strings)
	- [Extensions](#extensions)
	- [Recursing](#recursing)
	- [Multiple replacements per file](#multiple-replacements-per-file)
	- [Regex replacements](#regex-replacements)
	- [Advanced regex replacements](#advanced-regex-replacements)
	- [Numerical operations](#numerical-operations)
		- [Zero padding](#zero-padding)
	- [Verbosity](#verbosity)

## Installing go-rename

Simply run the following commands:

* `go get github.com/SlyMarbo/go-rename`
* `cd $GOPATH/src/github.com/SlyMarbo/go-rename`
* `go install`

Alternatively, use a pre-built binary:

- **OS X**:    [32-bit][osx_32], [64-bit][osx_64]
- **Linux**:   [32-bit][linux_32], [64-bit][linux_64]
- **Windows**: [32-bit][windows_32], [64-bit][windows_64]

[osx_32]: https://github.com/SlyMarbo/go-rename/blob/master/bin/osx_x86/go-rename?raw=true
[osx_64]: https://github.com/SlyMarbo/go-rename/blob/master/bin/osx_amd64/go-rename?raw=true
[linux_32]: https://github.com/SlyMarbo/go-rename/blob/master/bin/linux_x86/go-rename?raw=true
[linux_64]: https://github.com/SlyMarbo/go-rename/blob/master/bin/linux_amd64/go-rename?raw=true
[windows_32]: https://github.com/SlyMarbo/go-rename/blob/master/bin/windows_x86/go-rename?raw=true
[windows_64]: https://github.com/SlyMarbo/go-rename/blob/master/bin/windows_amd64/go-rename?raw=true

## Using go-rename

Renaming operations can be given a dry-run with the `-t` or `--test` flags, which print which operations
would have been performed without actually changing the filesystem.

Similarly, the `-l` or `--list` flag can be used to list which files would have been renamed without making
any changes.

### Simple strings

Simple renaming is performed with the `--from` and `--to` flags, which use plain strings. For example,

`$ go-rename --from "foo" --to "bar"`

will replace the first instance of `foo` with `bar` for each filename in the current directory.

### Extensions

To limit renaming to files with a particular extension, use the `--ext` flag. For example,

`$ go-rename --from "foo" --to "bar" --ext ".txt"`

will only perform the rename operation on files with the extension `.txt`.

### Recursing

To recurse into child directories, use the `-R` flag. For example,

`$ go-rename -R --from "foo" --to "bar"`

will replace the first instance of `foo` with `bar` for each filename in the current directory and any child
directories.

### Multiple replacements per file

To enable multiple replacements within each filename, use the `-n` flag. For example,

`$ go-rename -n 3 --from "foo" --to "bar"`

will replace up to 3 instances of `foo` with `bar` for each filename in the current directory.

To allow any number of replacements per file, use `-n -1`.

### Regex replacements

To use a regular expression in your matching string, use the `--from-regex` flag instead of `--from`. For
example,

`$ go-rename --from-regex "%d+" --to "num"`

will replace the first instance of one or more numbers with `num` for each filename in the current directory.

The syntax of the regular expressions accepted is the same general syntax used by Perl, Python, and other
languages. More precisely, it is the syntax accepted by RE2 and described at
http://code.google.com/p/re2/wiki/Syntax, except for \C. Furthermore, the backslash character (`\`) has been
replaced with the percent character (`%`) to avoid the need to escape backslashes in the shell. If in doubt
about how the regex will work, simply use `-t` to experiment.

### Advanced regex replacements

To use a regular expression in your replacement string, use the `--to-regex` flag instead of `--to`. For
example,

`$ go-rename --from-regex "(%d+)" --to-regex "%1_old"`

will append `_old` to the first instance of one or more numbers for each filename in the current directory.
Each group in the matching regex (groups are made with brackets) is assigned a number, starting at 1. The
content of a group can be inserted into the replacement regex by using `%#` where `#` is a group number.
Other than groupings, `--to-regex` is treated as a normal string. A literal percent character can be used
as `%%`, as with `--from-regex`.

Note that this only works when combined with `--from-regex`. If used with `--from`, it is treated as `--to`.

### Numerical operations

To performa a numerical operation on a number at the beginning or end of a filename (ignoring the extension),
use the `--number-start` or `--number-end` flag, as appropriate. For example,

`$ go-rename --number-start +2`

will modify each file in the current directory such that any number at the beginning of the filename is
incremented by two. `--number-start` and `--number-end` currently support addition and subtraction and both
will perform the renamings in such an order that it is safe (so `--number-start +1` would rename `2.jpg`
before `1.jpg`).

#### Zero padding

To ensure a certain number width with zero padding, use the `-z` or `--zero-pad` flag. For example,

`$ go-rename --number-start +2 -z 2`

will perform the numerical modification, and pad the result with zeros as necessary. In this case, `1.jpg`
would be renamed to `02.jpg`.

### Verbosity

Use the `-v` or `--verbose` flags to enable the printing of any warnings or other non-error information.
