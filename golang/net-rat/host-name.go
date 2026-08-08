//        ____()()     NetRat v0.3
//       /      @@     ~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"log/slog"
	"os"
)

type HostNameRat struct {
	name   string
	source string
}

func (rat *HostNameRat) Squeal(ni *NodeInfo) {
	slog.Debug("HostNameRat squeals:", "name", rat.name, "source", rat.source)
	ni.HostName = rat.name
}

func QueryHostNameRats(output chan<- Rat) {
	if host, err := os.Hostname(); err == nil {
		output <- &HostNameRat{
			source: "os.Hostname",
			name:   host,
		}
	}

}
