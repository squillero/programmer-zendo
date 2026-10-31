//        ____()()     NetRat v0.2.4
//       /      @@     ~~~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
)

// Map `key` to `info` with an explicit timestamp
func setAt[T comparable](e *Ephemeras[T], key string, info T, t time.Time) {
	e.TimedInfo[key] = TimedInfo[T]{Time: t, Info: info}
}

func TestEphemerasSetGet(t *testing.T) {
	e := MakeEphemeras[string]()
	if e.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", e.Len())
	}
	e.Set("a", "alpha")
	if got := e.Get("a"); got != "alpha" {
		t.Errorf("Get(a) = %q, want %q", got, "alpha")
	}
	if got := e.Get("missing"); got != "" {
		t.Errorf("Get(missing) = %q, want zero value", got)
	}
	if !e.HasKey("a") || e.HasKey("missing") {
		t.Errorf("HasKey misbehaves: %v", e.TimedInfo)
	}
	e.Set("a", "again")
	if e.Len() != 1 || e.Get("a") != "again" {
		t.Errorf("Set did not overwrite: %v", e.TimedInfo)
	}
}

func TestEphemerasSetZero(t *testing.T) {
	e := MakeEphemeras[GeoInfo]()
	e.Set("k", GeoInfo{City: "Turin", Country: "Italy"})
	e.SetZero("k")
	if !e.HasKey("k") || e.Get("k") != (GeoInfo{}) {
		t.Errorf("SetZero(k) left %v", e.Get("k"))
	}
}

func TestEphemerasCollections(t *testing.T) {
	e := MakeEphemeras[string]()
	e.Set("a", "x")
	e.Set("b", "y")
	e.Set("c", "x")

	keys := e.GetKeys("x")
	slices.Sort(keys)
	if !slices.Equal(keys, []string{"a", "c"}) {
		t.Errorf("GetKeys(x) = %v", keys)
	}
	if keys := e.GetKeys("z"); len(keys) != 0 {
		t.Errorf("GetKeys(z) = %v, want empty", keys)
	}
	all := e.GetAllKeys()
	slices.Sort(all)
	if !slices.Equal(all, []string{"a", "b", "c"}) {
		t.Errorf("GetAllKeys() = %v", all)
	}
	infos := e.GetAllInfos()
	slices.Sort(infos)
	if !slices.Equal(infos, []string{"x", "x", "y"}) {
		t.Errorf("GetAllInfos() = %v", infos)
	}
}

func TestEphemerasContent(t *testing.T) {
	e := MakeEphemeras[int]()
	e.Set("one", 1)
	e.Set("two", 2)
	e.Set("three", 3)

	sum := 0
	for _, v := range e.Content {
		sum += v
	}
	if sum != 6 {
		t.Errorf("sum over Content = %d, want 6", sum)
	}
	n := 0
	for range e.Content {
		n++
		break
	}
	if n != 1 {
		t.Errorf("break did not stop Content after %d items", n)
	}
}

func TestEphemerasGetLatestKey(t *testing.T) {
	e := MakeEphemeras[string]()
	if got := e.GetLatestKey(); got != "" {
		t.Errorf("GetLatestKey() on empty = %q, want \"\"", got)
	}
	now := time.Now()
	setAt(&e, "old", "", now.Add(-time.Hour))
	setAt(&e, "new", "", now)
	setAt(&e, "mid", "", now.Add(-time.Minute))
	if got := e.GetLatestKey(); got != "new" {
		t.Errorf("GetLatestKey() = %q, want %q", got, "new")
	}
}

func TestEphemerasInvalidate(t *testing.T) {
	e := MakeEphemeras[string]()
	now := time.Now()
	setAt(&e, "old1", "", now.Add(-2*time.Hour))
	setAt(&e, "old2", "", now.Add(-90*time.Minute))
	setAt(&e, "fresh", "", now)

	if n := e.Invalidate(now.Add(-time.Hour)); n != 2 {
		t.Errorf("Invalidate() = %d, want 2", n)
	}
	if !slices.Equal(e.GetAllKeys(), []string{"fresh"}) {
		t.Errorf("remaining keys = %v, want [fresh]", e.GetAllKeys())
	}
	if n := e.Invalidate(now.Add(-time.Hour)); n != 0 {
		t.Errorf("second Invalidate() = %d, want 0", n)
	}
}

func TestEphemerasJSONRoundTrip(t *testing.T) {
	e := MakeEphemeras[GeoInfo]()
	e.Set("1.2.3.4", GeoInfo{City: "Turin", Country: "Italy"})

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var back Ephemeras[GeoInfo]
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Get("1.2.3.4") != e.Get("1.2.3.4") {
		t.Errorf("round trip: got %v, want %v", back.Get("1.2.3.4"), e.Get("1.2.3.4"))
	}
	if !back.TimedInfo["1.2.3.4"].Time.Equal(e.TimedInfo["1.2.3.4"].Time) {
		t.Errorf("timestamp not preserved")
	}
}
