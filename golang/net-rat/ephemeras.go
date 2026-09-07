//        ____()()     NetRat v0.2.4
//       /      @@     ~~~~~~~~~~~~~
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

type TimedInfo[T any] struct {
	Time time.Time `json:"timestamp"`
	Info T         `json:"info"`
}

type Ephemeras[T any] struct {
	Content map[string]TimedInfo[T]
}

// Enable `for t, x := range test.Content` through rewiring (go1.23+)

func MakeEphemeras[T any]() Ephemeras[T] {
	return Ephemeras[T]{
		Content: make(map[string]TimedInfo[T]),
	}
}

// Number of ephemeral entries
func (e *Ephemeras[T]) Len() int {
	return len(e.Content)
}

// Add a value, update if newer
func (e *Ephemeras[T]) Add(key string, info T) {
	e.Content[key] = TimedInfo[T]{
		Info: info,
		Time: time.Now(),
	}
}

func (e *Ephemeras[T]) Invalidate(cutoff time.Time) int {
	num := 0

	for key, timedInfo := range e.Content {
		if timedInfo.Time.Before(cutoff) {
			delete(e.Content, key)
			slog.Debug("Deleting invalid ephemera:", "key", key, "ΔT", time.Since(timedInfo.Time))
			num++
		}
	}
	return num
}

func (e *Ephemeras[T]) Get(key string) T {
	if timedInfo, ok := e.Content[key]; ok {
		return timedInfo.Info
	} else {
		var zero T
		return zero
	}
}
