//        ____()()     NetRat v0.2.4
//       /      @@     ~~~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

const CACHE_FILENAME = "net-rat.json"

func cacheFile() string {
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, CACHE_FILENAME)
	} else {
		return "." + CACHE_FILENAME
	}
}

func DeleteCache() {
	fn := cacheFile()
	if err := os.Remove(fn); err != nil {
		slog.Debug("Error deleting cache:", "error", err.Error())
	} else {
		slog.Info("Deleting cache file:", "fn", fn)
	}
}

func SaveCache(ni *NodeInfo) error {
	fn := cacheFile()

	//fileData, err := json.MarshalIndent(*ni, "", "    ")
	fileData, err := json.Marshal(*ni)
	if err != nil {
		slog.Warn("Can't marshal data:.", "ni", ni)
		return err
	}
	if err = os.WriteFile(fn, fileData, 0644); err != nil {
		slog.Debug("Error writing cache:", "error", err.Error())
		return err
	}
	slog.Info("Cache saved:", "fn", fn)
	return nil
}

func LoadCache() *NodeInfo {
	fn := cacheFile()
	data, err := os.ReadFile(fn)
	if err != nil {
		slog.Debug("ReadFile failed:", "err", err.Error())
		return nil
	}
	var ni NodeInfo
	if err := json.Unmarshal(data, &ni); err != nil {
		slog.Debug("Error unmarshaling cache data:", "err", err.Error())
		return nil
	}
	if ni.SchemaVersion != RAT_VERSION {
		slog.Debug("Wrong cache version:", "actual", ni.SchemaVersion, "expected", RAT_VERSION)
		return nil
	}
	slog.Info("Cache loaded:", "fn", fn)
	return &ni
}
