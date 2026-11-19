// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2026 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"fmt"
	"strings"
)

// The db keys commands by their quoted text, so the quoting must be exactly
// the one of the shell version: bash's "printf %q" (bash 3.2, the /bin/bash of
// macOS, in a UTF-8 locale)

// shellQuote quotes s as bash's "printf %q" does
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	for i := 0; i < len(s); i++ {
		if isControl(s[i]) {
			return ansiQuote(s)
		}
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case ' ', '\t', '\n', '\'', '"', '\\', '|', '&', ';', '(', ')', '<', '>',
			'!', '{', '}', '*', '[', '?', ']', '^', '$', '`', ',':
			b.WriteByte('\\')
		case '#':
			if i == 0 {
				b.WriteByte('\\')
			}
		}
		b.WriteByte(c)
	}
	return b.String()
}

// ansiQuote quotes s in the $'...' form
func ansiQuote(s string) string {
	var b strings.Builder
	b.WriteString("$'")
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case 0x1b:
			b.WriteString(`\E`)
		case '\a':
			b.WriteString(`\a`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\v':
			b.WriteString(`\v`)
		case '\\', '\'':
			b.WriteByte('\\')
			b.WriteByte(c)
		default:
			if isControl(c) {
				fmt.Fprintf(&b, `\%03o`, c)
			} else {
				b.WriteByte(c)
			}
		}
	}
	b.WriteByte('\'')
	return b.String()
}

func isControl(c byte) bool {
	return c < 0x20 || c == 0x7f
}
