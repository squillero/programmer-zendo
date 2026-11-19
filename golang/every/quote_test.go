// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2026 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import "testing"

// The expected values are the output of "printf %q" of /bin/bash 3.2 on macOS
func TestShellQuote(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", "''"},
		{"plain", "plain"},
		{"a,b", `a\,b`},
		{"a=b", "a=b"},
		{"~x", "~x"},
		{"#a", `\#a`},
		{"a#b", "a#b"},
		{"a b", `a\ b`},
		{"it's", `it\'s`},
		{"!x^", `\!x\^`},
		{`a\b`, `a\\b`},
		{"{}", `\{\}`},
		{"é", "é"},
		{"+-./_:@%", "+-./_:@%"},
		{"a\nb", `$'a\nb'`},
		{"a\tb", `$'a\tb'`},
		{"a\rb", `$'a\rb'`},
		{"\x01", `$'\001'`},
		{"\x7f", `$'\177'`},
		{"\x1b", `$'\E'`},
		{"\a\b\f\v", `$'\a\b\f\v'`},
		{"a'b\nc", `$'a\'b\nc'`},
		{"\"x\n", `$'"x\n'`},
		{"a\\b\n", `$'a\\b\n'`},
		{"$x b\n", `$'$x b\n'`},
		{"é\n", `$'é\n'`},
	}
	for _, tt := range tests {
		if got := shellQuote(tt.in); got != tt.want {
			t.Errorf("shellQuote(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}
