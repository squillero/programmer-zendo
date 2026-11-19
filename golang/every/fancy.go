// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2026 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"fmt"
	"os"

	"github.com/muesli/termenv"
)

// Pretty printers: colors only when stdout is a terminal and NO_COLOR is unset

var (
	fancyProfile = termenv.Ascii
	fancyError   termenv.Color
	fancyWarning termenv.Color
	fancyInfo    termenv.Color
	fancyNote    termenv.Color
)

func setupFancy() {
	// The same colors as the shell version: tput's setaf 203, 208, 111 and 183
	// with 256 colors, otherwise setaf 1, 3, 4 and 5
	p := termenv.NewOutput(os.Stdout).EnvColorProfile()
	switch p {
	case termenv.Ascii:
		return
	case termenv.ANSI:
		fancyError, fancyWarning, fancyInfo, fancyNote = p.Color("1"), p.Color("3"), p.Color("4"), p.Color("5")
	default:
		fancyError, fancyWarning, fancyInfo, fancyNote = p.Color("203"), p.Color("208"), p.Color("111"), p.Color("183")
	}
	fancyProfile = p
}

// dot prints msg to stderr after a star of the given color
func dot(color termenv.Color, msg string) {
	if msg == "" {
		msg = "Mu"
	}
	star := "*"
	if color != nil {
		star = fancyProfile.String(star).Foreground(color).String()
	}
	fmt.Fprintln(os.Stderr, star+" "+msg)
}

func dotError(msg string)   { dot(fancyError, msg) }
func dotWarning(msg string) { dot(fancyWarning, msg) }
func dotInfo(msg string)    { dot(fancyInfo, msg) }
func dotNote(msg string)    { dot(fancyNote, msg) }
