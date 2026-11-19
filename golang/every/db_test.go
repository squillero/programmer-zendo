// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2026 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestParseEntry(t *testing.T) {
	const now = 1000
	valid := []string{"1000\tx", "0999\tx y", "1\t$'a\\nb'"}
	invalid := []string{"", "1000", "1001\tx", "x\ty", "1\t", "1\tx\ty", "-1\tx", " 1\tx", "99999999999999999999\tx"}
	for _, line := range valid {
		if _, _, ok := parseEntry(line, now); !ok {
			t.Errorf("parseEntry(%q) is not valid", line)
		}
	}
	for _, line := range invalid {
		if _, _, ok := parseEntry(line, now); ok {
			t.Errorf("parseEntry(%q) is valid", line)
		}
	}
}

func TestCheckDB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.tsv")
	if problems, ok := checkDB(path, 1000); !ok || problems != nil {
		t.Errorf("checkDB(missing) = %v %v, want valid", problems, ok)
	}
	db := "1\ta\n\n  \nx\tb\n2000\tc\n3\t\n1\t2\t3\n4\ta\nno tab"
	if err := os.WriteFile(path, []byte(db), 0o600); err != nil {
		t.Fatal(err)
	}
	want := []string{
		`line 4: timestamp "x" is not a number`,
		"line 5: timestamp 2000 is in the future",
		"line 6: empty command",
		"line 7: expected 2 tab-separated fields, found 3",
		"line 8: same command as line 1",
		"line 9: expected 2 tab-separated fields, found 1",
	}
	if problems, ok := checkDB(path, 1000); ok || !slices.Equal(problems, want) {
		t.Errorf("checkDB = %q %v, want %q", problems, ok, want)
	}
	if problems, ok := checkDB(dir, 1000); ok || len(problems) != 1 {
		t.Errorf("checkDB(dir) = %q %v, want not readable", problems, ok)
	}
}
