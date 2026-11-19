// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2026 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseInterval(t *testing.T) {
	tests := []struct {
		in    string
		scale int
		count int64
	}{
		{"30s", second, 30},
		{"2 minutes", minute, 2},
		{"24h", hour, 24},
		{"daily", day, 1},
		{"Day", day, 1},
		{"3 weeks", week, 3},
		{"month", month, 1},
		{"1h30m", minute, 90},
		{"1w2d", day, 9},
		{"1y6mo", month, 18},
		{"every 2 hours and 30 min", minute, 150},
		{"1h,30m", minute, 90},
		{"1h+30m", minute, 90},
		{"90 s 1 m", second, 150},
		{"007d", day, 7},
	}
	for _, tt := range tests {
		scale, count, ok := parseInterval(strings.Fields(tt.in))
		if !ok || scale != tt.scale || count != tt.count {
			t.Errorf("parseInterval(%q) = %d %d %v, want %d %d", tt.in, scale, count, ok, tt.scale, tt.count)
		}
	}
	for _, in := range []string{"", "2", "1 2 days", "fortnight", "1y2d", "1mo1w", "0s", "every", "d 3"} {
		if scale, count, ok := parseInterval(strings.Fields(in)); ok {
			t.Errorf("parseInterval(%q) = %d %d, want an error", in, scale, count)
		}
	}
}

func TestPeriods(t *testing.T) {
	rome, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		t.Skip("no time zone data")
	}
	saved := time.Local
	time.Local = rome
	t.Cleanup(func() { time.Local = saved })
	at := func(s string) int64 {
		tm, err := time.ParseInLocation("2006-01-02 15:04:05", s, rome)
		if err != nil {
			t.Fatal(err)
		}
		return tm.Unix()
	}
	tests := []struct {
		from, to string
		scale    int
		want     int64
	}{
		{"2026-11-19 10:00:59", "2026-11-19 10:01:00", minute, 1},
		{"2026-11-19 10:00:59", "2026-11-19 10:01:00", second, 1},
		{"2026-11-19 23:59:59", "2026-11-20 00:00:00", day, 1},
		{"2026-11-19 00:00:00", "2026-11-19 23:59:59", day, 0},
		{"2026-11-15 23:59:59", "2026-11-16 00:00:00", week, 1}, // Sunday to Monday
		{"2026-11-16 00:00:00", "2026-11-22 23:59:59", week, 0}, // Monday to Sunday
		{"2026-11-30 23:59:59", "2026-12-01 00:00:00", month, 1},
		{"2026-12-31 23:59:59", "2027-01-01 00:00:00", year, 1},
		{"2026-12-31 23:59:59", "2027-01-01 00:00:00", month, 1},
		{"2026-03-28 12:00:00", "2026-03-29 12:00:00", hour, 24}, // DST starts: wall-clock hours
		{"2026-03-28 12:00:00", "2026-03-29 12:00:00", second, 23 * 3600},
		{"2026-03-28 12:00:00", "2026-03-29 12:00:00", day, 1},
	}
	for _, tt := range tests {
		got := periods(at(tt.to))[tt.scale] - periods(at(tt.from))[tt.scale]
		if got != tt.want {
			t.Errorf("%s periods from %s to %s = %d, want %d", scales[tt.scale], tt.from, tt.to, got, tt.want)
		}
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		n    int64
		word string
		want string
	}{
		{0, "entry", "0 entries"},
		{1, "entry", "1 entry"},
		{2, "entry", "2 entries"},
		{1, "day", "1 day"},
		{3, "day", "3 days"},
	}
	for _, tt := range tests {
		if got := plural(tt.n, tt.word); got != tt.want {
			t.Errorf("plural(%d, %q) = %q, want %q", tt.n, tt.word, got, tt.want)
		}
	}
	for _, word := range scales {
		if got, want := plural(2, word), "2 "+word+"s"; got != want {
			t.Errorf("plural(2, %q) = %q, want %q", word, got, want)
		}
	}
}

func TestFormat(t *testing.T) {
	durations := map[int64]string{0: "0s", 59: "59s", 60: "1m", 3661: "1h 1m 1s", 90000: "1d 1h", 86400: "1d"}
	for s, want := range durations {
		if got := fmtDuration(s); got != want {
			t.Errorf("fmtDuration(%d) = %q, want %q", s, got, want)
		}
	}
	words := map[int64]string{0: "0 seconds", 1: "1 second", 61: "1 minute and 1 second",
		7200: "2 hours", 302400: "3 days and 12 hours", 86460: "1 day"}
	for s, want := range words {
		if got := fmtWords(s); got != want {
			t.Errorf("fmtWords(%d) = %q, want %q", s, got, want)
		}
	}
}
