// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2026 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

// Every runs a command only if enough time has passed since its last run, e.g.
//
//	every day : git pull          (once per calendar day)
//	every 24h : git pull          (when the clock hour has changed 24 times)
//	every 2 minutes : date        (2 minute changes: may be just 61 seconds apart)
//	every 120s : date             (at least 120 seconds apart)
//	every 2 weeks : brew cleanup  (calendar weeks, starting on Monday)
//	every 1h30m : make backup     (90 minute changes)
//	every day : 'cd ~/src && git pull && make'    (one argument: run by $SHELL)
//	every day : - <<'EOF'                         (a script on stdin, run by $SHELL)
//
// It is a port of the Bash script shell/every, and shares its db: runs are
// tracked per command in ${XDG_STATE_HOME:-~/.local/state}/every/last-run.tsv;
// entries not run in the last 10 years are pruned. The db is updated only when
// the command succeeds. By default only warnings are printed: -v also tells
// whether the command runs or is skipped, -q silences warnings too. With -a,
// a due command runs only if the user confirms within a few seconds.
// The db is locked only while it is being written, never while a command runs.
// The exit status is the command's when it runs, 1 when it doesn't (not due
// or not confirmed), and 2 on usage errors, so it can be tested with "if".
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/dustin/go-humanize/english"
)

const (
	lockTimeout = 2                                           // seconds to wait for another run to release the db
	checkMax    = 10                                          // problems shown by checkDB
	pruneYears  = 10                                          // entries not run for this long are pruned
	pruneAge    = (pruneYears*365 + (pruneYears+3)/4) * 86400 // in seconds, leap days included
	askTimeout  = 5                                           // seconds to wait for -a confirmation
)

var (
	prog    string
	quiet   bool
	verbose bool
	dryRun  bool
)

// usage returns the terse help of -h, or with full the help of -H, which adds
// the details to it
func usage(full bool) string {
	text := usageTerse
	if full {
		text += usageDetails
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return strings.NewReplacer("PROG", prog, "ASK_TIMEOUT", fmt.Sprint(askTimeout),
		"DEFAULT_SHELL", shell, "DB_PATH", dbPath, "PRUNE_SECONDS", fmt.Sprint(pruneAge), "PRUNE_YEARS", fmt.Sprint(pruneYears)).Replace(text)
}

const usageTerse = `Usage: PROG [-d] [-q|-v] [-a] [-t TAG] INTERVAL : COMMAND [ARGS...]
       PROG [-d] [-q|-v] [-a] [-t TAG] INTERVAL : 'SHELL CODE'
       PROG [-d] [-q|-v] [-a] [-t TAG] INTERVAL : - < SCRIPT
       PROG [-d] [-q|-v] [-c|-e]

Run COMMAND only if it has not run for INTERVAL; without arguments, list the
tracked commands. INTERVAL counts calendar periods in local time, e.g. "30s",
"2 minutes", "1h30m", "daily", "3 weeks", "1y6mo". A single argument with
shell characters, or "-" for a script on stdin, runs with $SHELL -c.
Successful runs are tracked in DB_PATH,
for PRUNE_SECONDS seconds (ie. PRUNE_YEARS years).

Options:
  -q  quiet: errors only
  -v  verbose: say what runs or is skipped, and why
  -a  ask before running; no answer in ASK_TIMEOUT seconds means no
  -d  dry run: change nothing; exit 0 if due, 1 if not
  -t  track the runs under TAG instead of the command
  -c  clear the db
  -e  edit the db with $EDITOR; with -d, only check the edits
  -h  show this help; -H show the full help

Exit status: the command's if it runs, 1 if not, 2 on usage errors.
`

const usageDetails = `
INTERVAL is a number of periods on one time scale: seconds, minutes, hours,
days, weeks, months or years, e.g. "30s", "2 minutes", "24h", "daily",
"3 weeks", "month". The command is due when the clock or calendar has moved
to a new period that many times since the last run (local time; weeks start
on Monday). So "2 minutes" is not "120 seconds": a run at 10:00:59 makes
"2 minutes" due at 10:02:00, but "120 seconds" only at 10:02:59. Likewise
"day" is due after midnight, "24h" after 24 changes of the hour.
  units: s(ec), m(in), h(our), d(ay), w(eek), mo(nth), y(ear),
         and hourly, daily, weekly, monthly, yearly
A combination is converted to the finest unit in it: "1h30m" is 90 minutes,
"1w2d" 9 days, "1y6mo" 18 months. Months and years can't be combined with
the other units, since their length varies.

COMMAND and its ARGS run as they are, with no shell: "find . -exec rm {} \;"
passes ";" to find. A single argument with spaces or shell characters is
shell code, run by $SHELL -c (DEFAULT_SHELL), so it can use ;, &&, |,
redirections and several lines. With "-", the script is read from stdin, e.g.
  PROG day : - <<'END'
  cd ~/src
  git pull && make
  END

The db has one line per command: the time of its last successful run (epoch)
TAB the command. A run that fails leaves it due. Entries not run in the last
PRUNE_SECONDS seconds (ie. PRUNE_YEARS years) are pruned, so longer intervals
can't be honored. The db is locked only while it is being written, never while
a command runs.

Options in detail:
  -q  print no warnings either, only errors; by default only warnings and
      errors are printed
  -v  say whether the command runs or is skipped, with db details and the
      periods passed since the last run on every time scale
  -a  when the command is due, ask before running it; no answer within
      ASK_TIMEOUT seconds means no, and the command stays due
  -d  neither run the command nor update the db
  -t  track the runs under TAG instead of the command itself, so a command
      can be changed without becoming due, and different commands can share
      one schedule, e.g. "PROG -t backup day : rsync ..."
  -c  empty the db, so every command is due on its next run
  -e  edit a copy of the db with $EDITOR; if the result is valid, it replaces
      the db, otherwise the db is left unchanged and the edits are kept in
      the copy, as it is if the db changed while it was being edited. With -d,
      only check the edited copy

Exit status: the command's when it runs; 1 when it doesn't, because it is not
due or was not confirmed; 2 on usage errors. With -d: 0 if the command is due,
1 if not. So "if PROG week : true; then ...; fi" runs once a week, and
"if PROG -d week : CMD; then ...; fi" tests whether CMD is due.
`

func say(msg string) {
	if verbose {
		dotInfo(msg)
	}
}

func sayVerbose(msg string) {
	if verbose {
		dotNote(msg)
	}
}

// exit releases the lock and removes temporary files before exiting
func exit(code int) {
	cleanup()
	os.Exit(code)
}

func die(msg string) {
	dotError(prog + ": " + msg)
	exit(2)
}

// shellChars are the characters that make a single argument shell code
var shellChars = regexp.MustCompile("[\\s;&|<>()$`\"'*?\\[\\]{}~\\\\]")

func main() {
	prog = filepath.Base(os.Args[0])
	setupFancy()
	setupPaths()

	// Options, as bash's getopts ":qvadcehHt:"
	clear, edit, ask, tagSet := false, false, false, false
	tag := ""
	args := os.Args[1:]
options:
	for len(args) > 0 {
		arg := args[0]
		if arg == "--" {
			args = args[1:]
			break
		}
		if len(arg) < 2 || arg[0] != '-' {
			break
		}
		args = args[1:]
		for j := 1; j < len(arg); j++ {
			switch arg[j] {
			case 'q':
				quiet, verbose = true, false
			case 'v':
				verbose, quiet = true, false
			case 'a':
				ask = true
			case 'd':
				dryRun = true
			case 'c':
				clear = true
			case 'e':
				edit = true
			case 't':
				switch {
				case j+1 < len(arg):
					tag = arg[j+1:]
				case len(args) > 0:
					tag, args = args[0], args[1:]
				default:
					die("-t needs an argument (see -h)")
				}
				tagSet = true
				continue options
			case 'h', 'H':
				fmt.Print(usage(arg[j] == 'H'))
				exit(0)
			default:
				fmt.Fprint(os.Stderr, usage(false))
				exit(2)
			}
		}
	}

	// The tag is the key of a db line: it can't hold the separators of the db
	if tagSet {
		if strings.TrimFunc(tag, unicode.IsSpace) == "" {
			die("empty tag")
		}
		if strings.ContainsAny(tag, "\t\n\r") {
			die("the tag can't contain tabs or newlines")
		}
		if len(args) == 0 || clear || edit {
			die("-t needs an interval and a command (see -h)")
		}
	}

	if edit {
		if len(args) > 0 || clear {
			die("-e takes no interval, command or -c (see -h)")
		}
		editDB()
		exit(0)
	}

	if clear {
		if len(args) > 0 {
			die("-c takes no interval or command (see -h)")
		}
		clearDB()
		exit(0)
	}

	if len(args) == 0 {
		listDB()
		exit(0)
	}

	var intervalWords []string
	for len(args) > 0 && args[0] != ":" {
		intervalWords = append(intervalWords, args[0])
		args = args[1:]
	}
	if len(args) == 0 {
		die("missing ':' between the interval and the command (see -h)")
	}
	args = args[1:]
	if len(intervalWords) == 0 {
		die("missing interval (see -h)")
	}
	if len(args) == 0 {
		die("missing command (see -h)")
	}

	kind, count, ok := parseInterval(intervalWords)
	if !ok {
		die(fmt.Sprintf("cannot understand interval '%s' (see -h)", strings.Join(intervalWords, " ")))
	}
	intervalText := plural(count, scales[kind])

	// Entries are purged after pruneAge, so longer intervals can't be honored
	maxCount := [...]int64{pruneAge, pruneAge / 60, pruneAge / 3600, pruneAge / 86400, pruneAge / 604800, 12 * pruneYears, pruneYears}
	if count > maxCount[kind] && !quiet {
		dotWarning(fmt.Sprintf("interval '%s' is longer than %d years, but entries are purged after %d days: the command may run sooner",
			intervalText, pruneYears, pruneAge/86400))
	}

	// A single argument that looks like shell code, or "-" for a script on stdin, is
	// run by $SHELL -c; anything else runs as is, so "find ... -exec rm {} \;" still
	// passes ";" to find. Stdin is read now, so it's consumed even if the run is skipped
	script := ""
	if len(args) == 1 {
		if args[0] == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				die("cannot read the script from stdin")
			}
			// As "$(cat)": no NUL bytes, no trailing newlines
			script = strings.TrimRight(strings.ReplaceAll(string(data), "\x00", ""), "\n")
			if strings.TrimFunc(script, unicode.IsSpace) == "" {
				die("empty script on stdin")
			}
		} else if strings.ContainsFunc(args[0], unicode.IsSpace) || shellChars.MatchString(args[0]) {
			script = args[0]
		}
	}

	now := time.Now().Unix()
	var key, cmdText string
	argv := args
	if script != "" {
		// The db has one entry per line: quote scripts with newlines or tabs
		key = script
		if strings.ContainsAny(script, "\n\t\r") {
			key = shellQuote(script)
		}
		// Show a multi-line script as its first non-blank line and its length
		nLines, first := 0, ""
		for _, line := range strings.Split(script, "\n") {
			if strings.TrimFunc(line, unicode.IsSpace) == "" {
				continue
			}
			if nLines++; nLines == 1 {
				first = strings.TrimLeftFunc(line, unicode.IsSpace)
			}
		}
		cmdText = first
		if nLines > 1 {
			cmdText = fmt.Sprintf("%s … (%d lines)", first, nLines)
		}
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		argv = []string{shell, "-c", script}
	} else {
		quoted := make([]string, len(args))
		for i, a := range args {
			quoted[i] = shellQuote(a)
		}
		cmdText = strings.Join(quoted, " ")
		key = cmdText
	}
	if tagSet {
		key = tag
		cmdText = fmt.Sprintf("%s [%s]", cmdText, tag)
	}

	// The latest valid run, if the command appears more than once
	last, found := int64(0), false
	entries := 0
	if isRegular(dbPath) {
		warnBadLines(now)
		if data, err := os.ReadFile(dbPath); err == nil {
			for _, line := range splitLines(string(data)) {
				if t, cmd, ok := parseEntry(line, now); ok && cmd == key && (!found || t > last) {
					last, found = t, true
				}
			}
			entries = strings.Count(string(data), "\n")
		}
	}

	sayVerbose(fmt.Sprintf("db: %s (%s)", dbPath, plural(int64(entries), "entry")))
	lastText := "never run"
	if found {
		lastText = fmt.Sprintf("last run %s ago", fmtDuration(now-last))
	}

	// Periods passed since the last run on each time scale, then how much is left
	// on the requested one before the command is due; empty if it is due now
	remaining := ""
	if found {
		lastPeriods, nowPeriods := periods(last), periods(now)
		var diffsText []string
		var passed int64
		for i := range scales {
			diff := nowPeriods[i] - lastPeriods[i]
			diffsText = append(diffsText, plural(diff, scales[i]))
			if i == kind {
				passed = diff
			}
		}
		sayVerbose("since the last run: " + strings.Join(diffsText, ", "))
		// Except for seconds, what's left is a number of period starts, not a duration
		if left := count - passed; left > 0 {
			switch {
			case kind == second:
				remaining = "due in " + fmtDuration(left)
			case left == 1:
				remaining = "due at the start of the next " + scales[kind]
			default:
				remaining = fmt.Sprintf("due when %d more %s have started", left, english.PluralWord(int(left), scales[kind], ""))
			}
		}
	}

	if remaining != "" {
		say(fmt.Sprintf("skipping %s (%s, %s)", cmdText, lastText, remaining))
		if !dryRun {
			updateDB(key, now, nil)
		}
		exit(1)
	}

	if dryRun {
		say(fmt.Sprintf("would run %s (%s)", cmdText, lastText))
		exit(0)
	}

	if ask && !confirm(cmdText, lastText) {
		say(fmt.Sprintf("skipping %s (%s, not confirmed)", cmdText, lastText))
		updateDB(key, now, nil)
		exit(1)
	}

	say(fmt.Sprintf("running %s (%s)", cmdText, lastText))
	status := runCommand(argv)
	if status == 0 {
		now = time.Now().Unix()
		updateDB(key, now, &now)
	} else if !quiet {
		dotWarning(fmt.Sprintf("%s failed with status %d; it will run again next time", cmdText, status))
	}
	exit(status)
}

// confirm asks whether to run the command; false on "no", timeout, or no terminal
func confirm(cmdText, lastText string) bool {
	fmt.Fprintf(os.Stderr, "%s (%s)? [y/N] ", cmdText, lastText)
	tty, err := os.Open("/dev/tty")
	if err != nil {
		fmt.Fprintln(os.Stderr)
		return false
	}
	answers := make(chan string, 1)
	go func() {
		line, err := bufio.NewReader(tty).ReadString('\n')
		if err != nil {
			close(answers)
			return
		}
		answers <- line
	}()
	select {
	case answer, ok := <-answers:
		if !ok {
			fmt.Fprintln(os.Stderr)
			return false
		}
		answer = strings.Trim(answer, " \t\n")
		return regexp.MustCompile(`^[Yy]([Ee][Ss])?$`).MatchString(answer)
	case <-time.After(askTimeout * time.Second):
		fmt.Fprintln(os.Stderr)
		return false
	}
}

// runCommand runs argv in the foreground and returns its exit status, as the
// shell would: 128+N if it is killed by signal N, 127 if it is not found, 126
// if it cannot be run. If the user interrupts it from the terminal, every dies
// of the same signal, as bash does
func runCommand(argv []string) int {
	cmd := exec.Command(argv[0], argv[1:]...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	// The terminal sends its signals to the command too: wait for it to
	// decide, instead of dying first
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, syscall.SIGINT, syscall.SIGQUIT)
	err := cmd.Run()
	signal.Stop(interrupts)

	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &exitErr):
		ws, ok := exitErr.Sys().(syscall.WaitStatus)
		if !ok || !ws.Signaled() {
			return exitErr.ExitCode()
		}
		select {
		case sig := <-interrupts:
			if sig == ws.Signal() {
				cleanup()
				signal.Reset(sig)
				syscall.Kill(os.Getpid(), ws.Signal())
				time.Sleep(time.Second)
			}
		default:
		}
		return 128 + int(ws.Signal())
	case errors.Is(err, exec.ErrNotFound):
		fmt.Fprintf(os.Stderr, "%s: %s: command not found\n", prog, argv[0])
		return 127
	case errors.Is(err, fs.ErrNotExist):
		fmt.Fprintf(os.Stderr, "%s: %s: No such file or directory\n", prog, argv[0])
		return 127
	default:
		fmt.Fprintf(os.Stderr, "%s: %s: %v\n", prog, argv[0], errors.Unwrap(err))
		return 126
	}
}
