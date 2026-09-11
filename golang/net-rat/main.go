//        ____()()     NetRat v0.2.4
//       /      @@     ~~~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const RAT_VERSION = "0.2.5"
const DEFAULT_TIMEOUT = 2
const CACHE_UPDATE_TIMEOUT = 10

var InvalidateEphemeras bool

func main() {
	log.SetPrefix("🐀 ") // 🐁 🐀
	log.SetFlags(log.Lmsgprefix + log.Lmicroseconds)

	verbosity := flag.Int("v", 0, "Verbosity, also -v or -vv")
	timeOut := flag.Int("t", -1, "Set timeout")
	cacheUpdate := flag.Bool("U", false, "Update cache and quit")
	printVersion := flag.Bool("V", false, "Display version info and quit")
	zapCache := flag.Bool("z", false, "Delete cache file")
	flag.BoolVar(&InvalidateEphemeras, "i", false, "Invalidate all ephemeras")
	var cmdline []string
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "-v") {
			v := 1 + strings.Count(a[2:], "v")
			cmdline = append(cmdline, "-v="+strconv.Itoa(v))
		} else {
			cmdline = append(cmdline, a)
		}
	}
	flag.CommandLine.Parse(cmdline)
	if *cacheUpdate {
		if *timeOut < 0 {
			*timeOut = CACHE_UPDATE_TIMEOUT
		}
	} else {
		if *timeOut < 0 {
			*timeOut = DEFAULT_TIMEOUT
		}
	}
	switch *verbosity + 1 {
	case 0:
		slog.SetLogLoggerLevel(slog.LevelError)
	case 1:
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case 2:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	default:
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	bannerLine1 := "NetRat v" + RAT_VERSION
	bannerLine2 := "(c) 2024-26 Giovanni Squillero / Politecnico di Torino"

	if *printVersion {
		os.Exit(0)
	} else if *verbosity >= 2 {
		log.Println(bannerLine1)
		log.Println(bannerLine2)
	}
	if *zapCache {
		DeleteCache()
	}
	var ni *NodeInfo
	if *cacheUpdate {
		to := time.Duration(*timeOut) * 1000 * time.Millisecond
		slog.Debug("main::Running in cache-update mode:", "pid", os.Getpid(), "ppid", os.Getppid())
		ni = UpdateCache(to)
	} else {
		to := time.Duration(*timeOut) * 1000 * time.Millisecond
		// Describe node!
		ni = DescribeNode(to)
	}
	if !*cacheUpdate {
		fmt.Println(ni.description)
	}
	if ni != nil {
		SaveCache(ni)
	}

	// Parent Execution Branch
	exe, err := os.Executable()
	if err != nil {
		slog.Error("os.Executable failed", "error", err)
		os.Exit(1)
	}
	cmd := exec.Command(exe, "-U")
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} // Detach process group (Unix)
	if err := cmd.Start(); err != nil {
		slog.Error("main: cmd.Start() failed", "cmd", cmd, "error", err)
		os.Exit(1)
	} else {
		slog.Info("main: spawned cache update")
	}
}
