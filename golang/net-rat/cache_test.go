//        ____()()     NetRat v0.2.4
//       /      @@     ~~~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Point os.UserCacheDir() to a fresh temporary directory
func useTempCache(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("LocalAppData", dir)
	if err := os.MkdirAll(filepath.Dir(cacheFile()), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestCacheRoundTrip(t *testing.T) {
	useTempCache(t)
	ni := makeTestNodeInfo(nodeSpec{
		host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}, egress: []string{"1.2.3.4"},
		geo: map[string]string{"1.2.3.4": "Italy"}, names: map[string]string{"1.2.3.4": "GARR"},
	})
	ni.IFaces.Set("aa:bb", "en0")
	ni.LinkLayerAddresses.Set("aa:bb", "10.0.0.2")

	if err := SaveCache(ni); err != nil {
		t.Fatal(err)
	}
	back := LoadCache()
	if back == nil {
		t.Fatal("LoadCache() = nil after SaveCache()")
	}
	if back.SchemaVersion != RAT_VERSION {
		t.Errorf("SchemaVersion = %q, want %q", back.SchemaVersion, RAT_VERSION)
	}
	if !back.Timestamp.Equal(ni.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", back.Timestamp, ni.Timestamp)
	}
	switch {
	case !back.HostName.HasKey("rat"):
		t.Errorf("HostName lost: %v", back.HostName.TimedInfo)
	case !back.IFaceAddresses.HasKey("10.0.0.2"):
		t.Errorf("IFaceAddresses lost: %v", back.IFaceAddresses.TimedInfo)
	case !back.PrivateAddresses.HasKey("10.0.0.2"):
		t.Errorf("PrivateAddresses lost: %v", back.PrivateAddresses.TimedInfo)
	case !back.EgressPoints.HasKey("1.2.3.4"):
		t.Errorf("EgressPoints lost: %v", back.EgressPoints.TimedInfo)
	case back.IFaces.Get("aa:bb") != "en0":
		t.Errorf("IFaces lost: %v", back.IFaces.TimedInfo)
	case back.LinkLayerAddresses.Get("aa:bb") != "10.0.0.2":
		t.Errorf("LinkLayerAddresses lost: %v", back.LinkLayerAddresses.TimedInfo)
	case back.Geo.Get("1.2.3.4") != (GeoInfo{Country: "Italy"}):
		t.Errorf("Geo lost: %v", back.Geo.TimedInfo)
	case back.NetworkName.Get("1.2.3.4") != "GARR":
		t.Errorf("NetworkName lost: %v", back.NetworkName.TimedInfo)
	}
}

func TestLoadCacheMissing(t *testing.T) {
	useTempCache(t)
	if ni := LoadCache(); ni != nil {
		t.Errorf("LoadCache() = %v, want nil", ni)
	}
}

func TestLoadCacheCorrupted(t *testing.T) {
	useTempCache(t)
	if err := os.WriteFile(cacheFile(), []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if ni := LoadCache(); ni != nil {
		t.Errorf("LoadCache() = %v, want nil", ni)
	}
}

func TestLoadCacheWrongSchema(t *testing.T) {
	useTempCache(t)
	ni := MakeNodeInfo()
	ni.SchemaVersion = "0.0.0"
	if err := SaveCache(ni); err != nil {
		t.Fatal(err)
	}
	if got := LoadCache(); got != nil {
		t.Errorf("LoadCache() = %v, want nil for schema %q", got, ni.SchemaVersion)
	}
}

func TestDeleteCache(t *testing.T) {
	useTempCache(t)
	if err := SaveCache(MakeNodeInfo()); err != nil {
		t.Fatal(err)
	}
	DeleteCache()
	if _, err := os.Stat(cacheFile()); !os.IsNotExist(err) {
		t.Errorf("cache file still there: %v", err)
	}
	DeleteCache() // Deleting a missing cache is harmless
}
