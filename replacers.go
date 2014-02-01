package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Replacer interface {
	Replace(string, string, *regexp.Regexp) string
}

type StringReplacer string

func (r StringReplacer) Replace(s, from string, regex *regexp.Regexp) string {
	if regex == nil {
		return strings.Replace(s, from, string(r), opts.Num)
	}

	to := new(bytes.Buffer)
	for _, match := range regex.FindAllString(s, opts.Num) {
		i := strings.Index(s, match)
		to.WriteString(s[:i])
		to.WriteString(string(r))
		s = s[i+len(match):]
	}

	return to.String()
}

type RegexReplacer struct {
	parts     []string
	groups    []int
	groupVals map[int]interface{}
}

func NewRegexReplacer(s string) *RegexReplacer {
	parts := make([]string, 0, 2)
	groups := make([]int, 0, 1)

	for {
		i := strings.Index(s, "%")
		if i < 0 {
			parts = append(parts, s)
			break
		}

		if len(s) == i+1 {
			complain("Error: Could not parse regex. % is not a valid combination.")
		}

		switch s[i+1] {
		case '%':
			parts = append(parts, s[:i+1])
			s = s[i+2:]
			continue

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			parts = append(parts, s[:i])

		default:
			complain("Error: Could not parse regex. " + s[i:i+2] + " is not a valid combination.")
		}

		group := 0
		i++
		for {
			if s[i] < '0' || s[i] > '9' || len(s) == i+1 {
				break
			}
			group = (group * 10) + int(s[i]-'0')
			i++
		}
		groups = append(groups, group)
	}

	r := new(RegexReplacer)
	r.parts = parts
	r.groups = groups
	r.groupVals = make(map[int]interface{})

	return r
}

func (r *RegexReplacer) Replace(s, _ string, regex *regexp.Regexp) string {
	length := len(r.parts)
	if len(r.groups) > length {
		length = len(r.groups)
	}
	to := new(bytes.Buffer)

	matchSets := regex.FindAllStringSubmatch(s, opts.Num)
	for _, matches := range matchSets {
		from := matches[0]
		i := strings.Index(s, from)
		to.WriteString(s[:i])
		s = s[i+len(from):]

		for i := range matches[1:] {
			r.groupVals[i+1] = matches[i+1]
		}

		buf := new(bytes.Buffer)
		for i := 0; i < length; i++ {
			if i < len(r.parts) {
				buf.WriteString(r.parts[i])
			}
			if i < len(r.groups) {
				fmt.Fprint(buf, r.groupVals[r.groups[i]])
			}
		}
		to.WriteString(buf.String())
	}

	return to.String()
}

type NumberReplacer struct {
	op    int
	start bool
}

func NewNumberReplacer(s string, start bool) *NumberReplacer {
	sign := 1
	switch s[0] {
	case '-':
		sign = -1
		fallthrough
	case '+':
		s = strings.TrimSpace(s[1:])
		fallthrough
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		i, err := strconv.Atoi(s)
		handle(err)
		return &NumberReplacer{sign * i, start}
	}

	complain("Error: Could not parse number op: " + s)
	return nil
}

func (n *NumberReplacer) Replace(s, _ string, regex *regexp.Regexp) string {
	match := regex.FindString(s)
	num, err := strconv.Atoi(match)
	handle(err)
	result := strconv.Itoa(num + n.op)
	if opts.ZeroPad > len(result) {
		zeros := strings.Repeat("0", opts.ZeroPad-len(result))
		result = zeros + result
	}

	if n.start {
		return strings.Replace(s, match, result, 1)
	}
	split := strings.LastIndex(s, match)
	return s[:split] + result + s[split+len(match):]
}
