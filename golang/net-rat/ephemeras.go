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
	"time"
)

type TimedInfo struct {
	Time time.Time `json:"timestamp"`
	Info string    `json:"info"`
}

type Ephemeras struct {
	Data []TimedInfo `json:"ephemeras"`
}

// Enable `for t, x := range test.Content` through rewiring (go1.23+)
func (e *Ephemeras) Content(yield func(time.Time, string) bool) {
	for _, s := range e.Data {
		if !yield(s.Time, s.Info) {
			return
		}
	}
}

// Number of ephemeral entries
func (e *Ephemeras) Len() int {
	return len(e.Data)
}

// Add a value, update if newer
func (e *Ephemeras) Add(info string) {
	for i := range e.Data {
		if e.Data[i].Info == info {
			e.Data[i].Time = time.Now()
			return
		}
	}
	e.Data = append(e.Data, TimedInfo{Info: info, Time: time.Now()})
}

func (e *Ephemeras) Invalidate(cutoff time.Time) int {
	num := 0
	var new []TimedInfo
	for _, d := range e.Data {
		if d.Time.After(cutoff) {
			new = append(new, d)
			slog.Debug("Found valid ephemera:", "val", d.Info, "ΔT", time.Since(d.Time))
		} else {
			slog.Debug("Deleting invalid ephemera:", "val", d.Info, "ΔT", time.Since(d.Time))
			num++
		}
	}
	e.Data = new
	return num
}

func (e *Ephemeras) Find(val string) time.Time {
	for _, d := range e.Data {
		if d.Info == val {
			return d.Time
		}
	}
	return time.Time{}
}

func (e *Ephemeras) Peek() string {
	return e.Data[0].Info
}
