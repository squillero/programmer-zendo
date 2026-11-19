// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2026 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
)

// The time scales, finest first
const (
	second = iota
	minute
	hour
	day
	week
	month
	year
)

var scales = [...]string{"second", "minute", "hour", "day", "week", "month", "year"}

// scaleSize is the length of each scale in its group's base unit: seconds for
// second..week, months for month and year
var scaleSize = [...]int64{1, 60, 3600, 86400, 604800, 1, 12}

var unitScale = map[string]int{
	"s": second, "sec": second, "secs": second, "second": second, "seconds": second,
	"m": minute, "min": minute, "mins": minute, "minute": minute, "minutes": minute,
	"h": hour, "hr": hour, "hrs": hour, "hour": hour, "hours": hour, "hourly": hour,
	"d": day, "day": day, "days": day, "daily": day,
	"w": week, "wk": week, "wks": week, "week": week, "weeks": week, "weekly": week,
	"mo": month, "mon": month, "month": month, "months": month, "monthly": month,
	"y": year, "yr": year, "yrs": year, "year": year, "years": year, "yearly": year, "annually": year,
}

var (
	digitLetter = regexp.MustCompile(`([0-9])([a-z])`)
	letterDigit = regexp.MustCompile(`([a-z])([0-9])`)
	digits      = regexp.MustCompile(`^[0-9]+$`)
)

// parseInterval parses a human-readable interval into a time scale and a
// count, e.g. "1h30m" into minute and 90
func parseInterval(words []string) (scale int, count int64, ok bool) {
	text := strings.ToLower(strings.Join(words, " ") + " ")
	text = digitLetter.ReplaceAllString(text, "$1 $2")
	text = letterDigit.ReplaceAllString(text, "$1 $2")
	text = strings.NewReplacer(",", " ", "+", " ").Replace(text)
	n := int64(-1) // the pending number, if any
	finest := -1
	var total int64
	for _, token := range strings.FieldsFunc(text, isIFS) {
		if digits.MatchString(token) {
			if n >= 0 {
				return 0, 0, false
			}
			v, err := strconv.ParseInt(token, 10, 64)
			if err != nil {
				return 0, 0, false
			}
			n = v
			continue
		}
		if token == "and" || token == "every" {
			continue
		}
		i, known := unitScale[token]
		if !known {
			return 0, 0, false
		}
		// Months and years have no fixed length in seconds: don't mix the groups
		if finest >= 0 && (finest >= month) != (i >= month) {
			return 0, 0, false
		}
		if finest < 0 || i < finest {
			finest = i
		}
		if n < 0 {
			n = 1
		}
		total += n * scaleSize[i]
		n = -1
	}
	if n >= 0 || finest < 0 {
		return 0, 0, false
	}
	count = total / scaleSize[finest]
	if count <= 0 {
		return 0, 0, false
	}
	return finest, count, true
}

// periods returns the number of the period that contains epoch t on each time
// scale, in local time
func periods(t int64) [len(scales)]int64 {
	lt := time.Unix(t, 0).Local()
	_, offset := lt.Zone()
	local := t + int64(offset)
	d := local / 86400
	y := int64(lt.Year())
	// 1970-01-01 was a Thursday: shift by 3 days so weeks start on Monday
	return [len(scales)]int64{t, local / 60, local / 3600, d, (d + 3) / 7, y*12 + int64(lt.Month()) - 1, y}
}

// fmtDuration formats seconds as e.g. "2d 3h 5m"
func fmtDuration(s int64) string {
	var parts []string
	for _, u := range []struct {
		size int64
		name string
	}{{86400, "d"}, {3600, "h"}, {60, "m"}} {
		if s >= u.size {
			parts = append(parts, strconv.FormatInt(s/u.size, 10)+u.name)
			s %= u.size
		}
	}
	if s > 0 || len(parts) == 0 {
		parts = append(parts, strconv.FormatInt(s, 10)+"s")
	}
	return strings.Join(parts, " ")
}

// fmtWords formats seconds in words, with the two largest units, e.g. "3 days and 12 hours"
func fmtWords(s int64) string {
	sizes := []int64{86400, 3600, 60, 1}
	names := []string{"day", "hour", "minute", "second"}
	i := 0
	for i < 3 && s < sizes[i] {
		i++
	}
	out := plural(s/sizes[i], names[i])
	if i < 3 {
		if n := s % sizes[i] / sizes[i+1]; n > 0 {
			out += " and " + plural(n, names[i+1])
		}
	}
	return out
}

func fmtDate(t int64) string {
	return time.Unix(t, 0).Local().Format("2006-01-02 15:04:05")
}

// plural formats count n with word, in the plural unless n is 1, e.g. "2 entries"
func plural(n int64, word string) string {
	return fmt.Sprintf("%d %s", n, english.PluralWord(int(n), word, ""))
}

// isIFS tells whether r splits words, as bash's default IFS
func isIFS(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}
