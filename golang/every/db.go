// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2026 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"bytes"
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// The db is a TSV file with one entry per line: last run (epoch) TAB command.
// It is shared with the shell version, so its format, location and lock must
// stay the same

var (
	dbDir    string
	dbPath   string
	lockPath string
	locked   bool   // whether this run holds the lock
	editDir  string // the temporary directory of -e, removed on exit
)

func setupPaths() {
	state := os.Getenv("XDG_STATE_HOME")
	if state == "" {
		state = os.Getenv("HOME") + "/.local/state"
	}
	dbDir = state + "/every"
	dbPath = dbDir + "/last-run.tsv"
	lockPath = dbPath + ".lock"
}

// cleanup releases the lock and removes the temporary directory of -e, if any
func cleanup() {
	if locked {
		os.RemoveAll(lockPath)
	}
	if editDir != "" {
		os.RemoveAll(editDir)
	}
}

// splitLines splits text into lines, as awk reads them
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// isBlank tells whether line is empty or only white space
func isBlank(line string) bool {
	return strings.TrimLeft(line, " \t\n\v\f\r") == ""
}

// parseEntry parses a db line; ok is false if it isn't a valid entry, with a
// timestamp not after now
func parseEntry(line string, now int64) (stamp int64, cmd string, ok bool) {
	fields := strings.Split(line, "\t")
	if len(fields) != 2 || !digits.MatchString(fields[0]) || fields[1] == "" {
		return 0, "", false
	}
	stamp, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || stamp > now {
		return 0, "", false
	}
	return stamp, fields[1], true
}

func isRegular(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsRegular()
}

// readDB returns the lines of the db; a missing db has none
func readDB() ([]string, error) {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return nil, err
	}
	return splitLines(string(data)), nil
}

// lockDB locks the db while it is being written: call unlockDB as soon as it
// is done (cleanup also releases it, if the run dies meanwhile). The lock is a
// directory holding the pid of its owner, to tell a stale lock
func lockDB() {
	if err := os.MkdirAll(dbDir, 0o777); err != nil {
		die("cannot create " + dbDir)
	}
	for tries := 0; os.Mkdir(lockPath, 0o777) != nil; {
		owner := ""
		if data, err := os.ReadFile(lockPath + "/pid"); err == nil {
			owner = strings.TrimRight(string(data), "\n")
		}
		if owner != "" && !isAlive(owner) {
			die(fmt.Sprintf("cannot lock %s: stale lock of pid %s, which is gone; remove %s", dbPath, owner, lockPath))
		}
		if owner == "" {
			owner = "?"
		}
		if tries == 5 {
			sayVerbose(fmt.Sprintf("waiting for %s, in use by pid %s", dbPath, owner))
		}
		if tries++; tries > lockTimeout*5 {
			die(fmt.Sprintf("cannot lock %s: in use by pid %s (remove %s if stale)", dbPath, owner, lockPath))
		}
		time.Sleep(200 * time.Millisecond)
	}
	locked = true
	os.WriteFile(lockPath+"/pid", []byte(strconv.Itoa(os.Getpid())+"\n"), 0o666)
}

// isAlive tells whether the process pid exists, like "kill -0"
func isAlive(pid string) bool {
	n, err := strconv.Atoi(pid)
	return err == nil && syscall.Kill(n, 0) == nil
}

func unlockDB() {
	os.RemoveAll(lockPath)
	locked = false
}

// replaceDB atomically replaces the db with lines; the db must be locked
func replaceDB(lines []string) {
	tmp, err := os.CreateTemp(dbDir, filepath.Base(dbPath)+".*")
	if err != nil {
		die("cannot write to " + dbDir)
	}
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(line + "\n")
	}
	_, err = tmp.WriteString(b.String())
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), dbPath)
	}
	if err != nil {
		os.Remove(tmp.Name())
		die("cannot update " + dbPath)
	}
}

// updateDB rewrites the db dropping stale entries; with stamp, it also sets
// key's last run to it. Blank lines are dropped; other bad lines (see checkDB)
// are kept as they are, for the user to fix with -e
func updateDB(key string, now int64, stamp *int64) {
	lockDB()
	if f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_WRONLY, 0o666); err == nil {
		f.Close()
	}
	lines, err := readDB()
	if err != nil {
		die("cannot update " + dbPath)
	}
	var out []string
	for _, line := range lines {
		if isBlank(line) {
			continue
		}
		t, cmd, ok := parseEntry(line, now)
		if !ok {
			out = append(out, line)
			continue
		}
		if (stamp != nil && cmd == key) || t < now-pruneAge {
			continue
		}
		out = append(out, line)
	}
	if stamp != nil {
		out = append(out, strconv.FormatInt(*stamp, 10)+"\t"+key)
	}
	replaceDB(out)
	unlockDB()
}

// checkDB checks that the db, or an edited copy of it, is valid; it returns
// one line per problem (at most checkMax). A missing or empty db is valid:
// nothing has run yet. Blank lines are ignored
func checkDB(path string, now int64) (problems []string, ok bool) {
	if _, err := os.Stat(path); err != nil {
		return nil, true
	}
	data, err := os.ReadFile(path)
	if !isRegular(path) || err != nil {
		return []string{"not a readable file"}, false
	}
	n := 0
	bad := func(nr int, why string) {
		if n++; n <= checkMax {
			problems = append(problems, fmt.Sprintf("line %d: %s", nr, why))
		}
	}
	seen := map[string]int{}
	for i, line := range splitLines(string(data)) {
		nr := i + 1
		if isBlank(line) {
			continue
		}
		fields := strings.Split(line, "\t")
		switch {
		case len(fields) != 2:
			bad(nr, fmt.Sprintf("expected 2 tab-separated fields, found %d", len(fields)))
		case !digits.MatchString(fields[0]):
			bad(nr, fmt.Sprintf("timestamp %q is not a number", fields[0]))
		case isFuture(fields[0], now):
			bad(nr, fmt.Sprintf("timestamp %s is in the future", fields[0]))
		case fields[1] == "":
			bad(nr, "empty command")
		case seen[fields[1]] > 0:
			bad(nr, fmt.Sprintf("same command as line %d", seen[fields[1]]))
		default:
			seen[fields[1]] = nr
		}
	}
	if n > checkMax {
		problems = append(problems, fmt.Sprintf("... and %d more", n-checkMax))
	}
	return problems, n == 0
}

// isFuture tells whether the timestamp (all digits) is after now
func isFuture(stamp string, now int64) bool {
	t, err := strconv.ParseInt(stamp, 10, 64)
	return err != nil || t > now
}

// warnBadLines warns about bad lines in the db, unless quiet. They are
// ignored, but kept
func warnBadLines(now int64) {
	if quiet {
		return
	}
	problems, ok := checkDB(dbPath, now)
	if ok {
		return
	}
	dotWarning(fmt.Sprintf("%s: bad lines in %s are ignored; fix them with %s -e", prog, dbPath, prog))
	printIndented(problems)
}

func printIndented(lines []string) {
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, "    "+line)
	}
}

// listDB prints the tracked commands, most recent first
func listDB() {
	now := time.Now().Unix()
	type entry struct {
		stamp int64
		line  string
	}
	var entries []entry
	if isRegular(dbPath) {
		warnBadLines(now)
		lines, _ := readDB()
		for _, line := range lines {
			if t, _, ok := parseEntry(line, now); ok {
				entries = append(entries, entry{t, line})
			}
		}
	}
	// As "sort -rn": by timestamp, then by the whole line, both reversed
	slices.SortStableFunc(entries, func(a, b entry) int {
		if c := cmp.Compare(b.stamp, a.stamp); c != 0 {
			return c
		}
		return strings.Compare(b.line, a.line)
	})
	var out []string
	for _, e := range entries {
		if now-e.stamp > pruneAge {
			continue
		}
		_, cmd, _ := parseEntry(e.line, now)
		out = append(out, fmt.Sprintf("%s (%s ago): %s", fmtDate(e.stamp), fmtWords(now-e.stamp), cmd))
	}
	sayVerbose(fmt.Sprintf("db: %s (%s)", dbPath, plural(int64(len(out)), "entry")))
	if len(out) > 0 {
		fmt.Println(strings.Join(out, "\n"))
	} else if !quiet {
		dotWarning("no commands tracked yet")
	}
}

// clearDB forgets all tracked commands
func clearDB() {
	count := 0
	if isRegular(dbPath) {
		lines, _ := readDB()
		for _, line := range lines {
			if strings.Count(line, "\t") == 1 {
				count++
			}
		}
	}
	entries := plural(int64(count), "entry")
	sayVerbose("db: " + dbPath)
	if dryRun {
		if !quiet {
			dotInfo("would clear " + entries + " from the database")
		}
		return
	}
	if isRegular(dbPath) {
		lockDB()
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			die("cannot remove " + dbPath)
		}
		unlockDB()
	}
	if !quiet {
		dotInfo("cleared " + entries + " from the database")
	}
}

// editDB edits a copy of the db with $EDITOR, and replaces the db with it only
// if it is valid and the db did not change while the editor was open. The db
// is locked only to replace it, so other runs are not blocked by the editor
func editDB() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		die("EDITOR is not set")
	}
	tmpDir := os.Getenv("TMPDIR")
	if tmpDir == "" {
		tmpDir = "/tmp"
	}
	dir, err := os.MkdirTemp(tmpDir, "every.")
	if err != nil {
		die("cannot create a temporary directory")
	}
	// The same path mktemp would print, even if TMPDIR ends with a slash
	editDir = tmpDir + "/" + filepath.Base(dir)
	copyPath := editDir + "/last-run.tsv"
	var orig []byte
	if isRegular(dbPath) {
		if orig, err = os.ReadFile(dbPath); err != nil {
			die("cannot copy " + dbPath)
		}
	} else if _, err := os.Lstat(dbPath); err == nil {
		die(dbPath + " is not a regular file")
	}
	if err := os.WriteFile(editDir+"/original.tsv", orig, 0o600); err != nil {
		die("cannot copy " + dbPath)
	}
	if err := os.WriteFile(copyPath, orig, 0o600); err != nil {
		die("cannot copy " + dbPath)
	}
	sayVerbose(fmt.Sprintf("db: %s, editing %s", dbPath, copyPath))

	// EDITOR may hold options, e.g. "emacs -nw": split it into words
	argv := append(strings.FieldsFunc(editor, isIFS), copyPath)
	if runCommand(argv) != 0 {
		if !quiet {
			dotWarning(fmt.Sprintf("%s: %s failed: %s left unchanged", prog, editor, dbPath))
		}
		exit(1)
	}
	edited, err := os.ReadFile(copyPath)
	if err != nil {
		die("cannot update " + dbPath)
	}
	if bytes.Equal(orig, edited) {
		say(fmt.Sprintf("no changes: %s left unchanged", dbPath))
		exit(0)
	}

	if problems, ok := checkDB(copyPath, time.Now().Unix()); !ok {
		// Keep the edited copy, so the changes are not lost
		if !quiet {
			dotWarning(fmt.Sprintf("%s: the edited db is not valid: %s left unchanged, edits kept in %s", prog, dbPath, copyPath))
			printIndented(problems)
		}
		editDir = ""
		exit(1)
	}
	if dryRun {
		say(fmt.Sprintf("the edited db is valid, but %s is left unchanged (dry run)", dbPath))
		exit(0)
	}

	lockDB()
	current, err := os.ReadFile(dbPath)
	if os.IsNotExist(err) {
		current, err = nil, nil
	}
	if err != nil || !bytes.Equal(orig, current) {
		unlockDB()
		if !quiet {
			dotWarning(fmt.Sprintf("%s: %s changed while it was being edited: left unchanged, edits kept in %s", prog, dbPath, copyPath))
		}
		editDir = ""
		exit(1)
	}
	// Drop blank lines: deleting every line in an editor often leaves one
	var lines []string
	for _, line := range splitLines(string(edited)) {
		if !isBlank(line) {
			lines = append(lines, line)
		}
	}
	replaceDB(lines)
	unlockDB()
	say(dbPath + " updated")
}
