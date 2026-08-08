//        ____()()     NetRat v0.3
//       /      @@     ~~~~~~~~~~~
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
	"strconv"
	"strings"
	"time"
)

const RAT_VERSION = "0.3"
const DEFAULT_TIMEOUT = 2

type verbosityLevel int

var InvalidateEphemeras bool
var GlobalTimeout time.Duration

func main() {
	log.SetPrefix("🐀 ") // 🐁 🐀
	log.SetFlags(log.Lmsgprefix + log.Lmicroseconds)

	printVersion := flag.Bool("V", false, "Display version info and quit")
	verbosity := flag.Int("v", 1, "Verbosity, also -v or -vv")
	zapCache := flag.Bool("z", false, "Delete cache file")
	waitForever := flag.Bool("w", false, "Wait indefinitely until connected")
	timeOut := flag.Int("t", DEFAULT_TIMEOUT, "Set timeout")
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
	bannerLine2 := "(c) 2024-26 Giovananni Squillero <giovanni.squillero@polito.it>"

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
	if *waitForever {
		InvalidateEphemeras = true
		GlobalTimeout = 24 * time.Hour
	} else {
		GlobalTimeout = time.Duration(*timeOut) * time.Second
	}

	DescribeNode()
}
