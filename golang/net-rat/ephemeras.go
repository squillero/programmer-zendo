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

type Ephemeras[T comparable] struct {
	TimedInfo map[string]TimedInfo[T]
}

// Enable `for key, info := range data.Content` through rewiring (go1.23+)
func (e *Ephemeras[T]) Content(yield func(string, T) bool) {
	for k, ti := range e.TimedInfo {
		if !yield(k, ti.Info) {
			return
		}
	}
}

func MakeEphemeras[T comparable]() Ephemeras[T] {
	return Ephemeras[T]{
		TimedInfo: make(map[string]TimedInfo[T]),
	}
}

// Number of ephemeral entries
func (e *Ephemeras[T]) Len() int {
	return len(e.TimedInfo)
}

// Map `key` to `info` and set/update Time
func (e *Ephemeras[T]) Set(key string, info T) {
	e.TimedInfo[key] = TimedInfo[T]{
		Info: info,
		Time: time.Now(),
	}
}

// Map `key` to `info` and set/update Time
func (e *Ephemeras[T]) SetZero(key string) {
	var zero T
	e.TimedInfo[key] = TimedInfo[T]{
		Info: zero,
		Time: time.Now(),
	}
}

// All possible getters:
// Get `info` from `key`. Return `zero T` if not existing.
func (e *Ephemeras[T]) Get(key string) T {
	if timedInfo, ok := e.TimedInfo[key]; ok {
		return timedInfo.Info
	} else {
		var zero T
		return zero
	}
}

// Get all `key` mapped to `info`. Return `zero T` if not existing.
func (e *Ephemeras[T]) GetKeys(info T) []string {
	keys := make([]string, 0)
	for k, i := range e.TimedInfo {
		if info == i.Info {
			keys = append(keys, k)
		}
	}
	return keys
}

// Get all `info`.
func (e *Ephemeras[T]) GetAllInfos() []T {
	all := make([]T, 0)
	for _, i := range e.TimedInfo {
		all = append(all, i.Info)
	}
	return all
}

// Get all `key`.
func (e *Ephemeras[T]) GetAllKeys() []string {
	all := make([]string, 0)
	for k := range e.TimedInfo {
		all = append(all, k)
	}
	return all
}

// Get the latest `key`.
func (e *Ephemeras[T]) GetLatestKey() string {
	var zero T
	latestKey := ""
	latestInfo := TimedInfo[T]{
		Time: time.Time{},
		Info: zero,
	}

	for k, i := range e.TimedInfo {
		if i.Time.After(latestInfo.Time) {
			latestInfo = i
			latestKey = k
		}
	}
	return latestKey
}

// Check if `key` exists.
func (e *Ephemeras[T]) HasKey(key string) bool {
	_, ok := e.TimedInfo[key]
	return ok
}

// Invalidate all `info` older than `cutoff`.
func (e *Ephemeras[T]) Invalidate(cutoff time.Time) int {
	num := 0

	for key, timedInfo := range e.TimedInfo {
		if timedInfo.Time.Before(cutoff) {
			delete(e.TimedInfo, key)
			slog.Debug("Deleting outdated ephemera:", "key", key, "ΔT", time.Since(timedInfo.Time))
			slog.Error("XXXX:", "key", key, "T", timedInfo.Time, "CT", cutoff, "ΔT", time.Since(timedInfo.Time))
			num++
		}
	}
	return num
}
