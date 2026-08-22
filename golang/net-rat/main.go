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

const RAT_VERSION = "0.2.4"
const DEFAULT_TIMEOUT = 2
const CACHE_UPDATE_TIMEOUT = 10

var InvalidateEphemeras bool

func main() {
	log.SetPrefix("🐀 ") // 🐁 🐀
	log.SetFlags(log.Lmsgprefix + log.Lmicroseconds)

	verbosity := flag.Int("v", -1, "Verbosity, also -v or -vv")
	timeOut := flag.Int("t", -1, "Set timeout")
	cacheUpdate := flag.Bool("U", false, "Update cache and quit")
	printVersion := flag.Bool("V", false, "Display version info and quit")
	zapCache := flag.Bool("z", false, "Delete cache file")
	flag.BoolVar(&InvalidateEphemeras, "i", false, "Invalidate all ephemeras")
	var cmdline []string
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "-v") {
			v := 2 + strings.Count(a[2:], "v")
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
		if *verbosity < 0 {
			*verbosity = 0
		}
	} else {
		if *timeOut < 0 {
			*timeOut = DEFAULT_TIMEOUT
		}
		if *verbosity < 0 {
			*verbosity = 1
		}
	}
	switch *verbosity {
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
		fmt.Println(bannerLine1 + " " + bannerLine2)
		os.Exit(0)
	} else if *verbosity >= 2 {
		log.Println(bannerLine1)
		log.Println(bannerLine2)
	}
	if *zapCache {
		DeleteCache()
	}
	t1 := time.Duration(*timeOut) * time.Second
	t2 := t1 / 20
	if *cacheUpdate {
		slog.Debug("Running in cache-update mode", "pid", os.Getgid())
		t2 = -1
	}

	// Describe node!
	// ni := NodeInfo{Timestamp: time.Now()}
	ni := DescribeNode(t1, t2)
	fmt.Println(ni.NetworkName)
	SaveCache(ni)
	if *cacheUpdate {
		slog.Info("Cache update completed")
		os.Exit(0)
	}

	fmt.Println(ni.description)

	// Parent Execution Branch
	exe, err := os.Executable()
	if err != nil {
		slog.Error("os.Executable failed", "error", err)
		os.Exit(1)
	}
	cmd := exec.Command(exe, "-U", "-v")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} // Detach process group (Unix)
	if err := cmd.Start(); err != nil {
		slog.Error("cmd.Start() failed", "cmd", cmd, "error", err)
		os.Exit(1)
	}
}
