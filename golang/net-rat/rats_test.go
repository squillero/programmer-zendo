//        ____()()     NetRat v0.2.4
//       /      @@     ~~~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCanonize(t *testing.T) {
	tests := map[string]string{
		"IT":                       "Italy",
		"it":                       "Italy",
		"ITA":                      "Italy",
		"U.S.A.":                   "United States",
		"United States of America": "United States",
		"Italy":                    "Italy",
		"Atlantis":                 "Atlantis",
		"":                         "",
	}
	for in, want := range tests {
		if got := canonize(in); got != want {
			t.Errorf("canonize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSimpleRatsSqueal(t *testing.T) {
	ni := MakeNodeInfo()
	for _, rat := range []Rat{
		&HostNameRat{name: "rat", source: "test"},
		&IFaceRat{ip: "10.0.0.2"},
		&IFaceRat{ip: "127.0.0.1", loopback: true},
		&PrivateRat{ip: "10.8.0.2", source: "test"},
		&PrivateRat{ip: "127.0.0.1", source: "test", loopback: true},
		&GlobalRat{ip: "1.2.3.4", source: "test"},
		&HwRat{iface: "en0", ip: "10.0.0.2", hw: "aa:bb"},
		&HwRat{iface: "lo0", ip: "127.0.0.1", hw: ""},
		&RDAPRat{ip: "1.2.3.4", name: "GARR", country: "IT"},
	} {
		rat.Squeal(ni)
	}

	switch {
	case ni.HostName.Len() != 1 || !ni.HostName.HasKey("rat"):
		t.Errorf("HostName = %v", ni.HostName.TimedInfo)
	case ni.IFaceAddresses.Len() != 1 || !ni.IFaceAddresses.HasKey("10.0.0.2"):
		t.Errorf("IFaceAddresses = %v", ni.IFaceAddresses.TimedInfo)
	case ni.PrivateAddresses.Len() != 1 || !ni.PrivateAddresses.HasKey("10.8.0.2"):
		t.Errorf("PrivateAddresses = %v", ni.PrivateAddresses.TimedInfo)
	case ni.EgressPoints.Len() != 1 || !ni.EgressPoints.HasKey("1.2.3.4"):
		t.Errorf("EgressPoints = %v", ni.EgressPoints.TimedInfo)
	case ni.LinkLayerAddresses.Len() != 1 || ni.LinkLayerAddresses.Get("aa:bb") != "10.0.0.2":
		t.Errorf("LinkLayerAddresses = %v", ni.LinkLayerAddresses.TimedInfo)
	case ni.IFaces.Len() != 1 || ni.IFaces.Get("aa:bb") != "en0":
		t.Errorf("IFaces = %v", ni.IFaces.TimedInfo)
	case ni.NetworkName.Get("1.2.3.4") != "GARR":
		t.Errorf("NetworkName = %v", ni.NetworkName.TimedInfo)
	}
}

func TestHwRatWithoutIFacePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("HwRat with no interface did not panic")
		}
	}()
	(&HwRat{ip: "10.0.0.2", hw: "aa:bb"}).Squeal(MakeNodeInfo())
}

func TestGeoRatSqueal(t *testing.T) {
	tests := []struct {
		name string
		old  *GeoInfo
		rat  GeoRat
		want GeoInfo
	}{
		{
			name: "fresh record is canonized",
			rat:  GeoRat{ip: "1.2.3.4", city: "Turin", country: "IT"},
			want: GeoInfo{City: "Turin", Country: "Italy"},
		},
		{
			name: "missing city is patched",
			old:  &GeoInfo{City: "Turin", Country: "Italy"},
			rat:  GeoRat{ip: "1.2.3.4", country: "Italy"},
			want: GeoInfo{City: "Turin", Country: "Italy"},
		},
		{
			name: "missing country is patched",
			old:  &GeoInfo{City: "Turin", Country: "Italy"},
			rat:  GeoRat{ip: "1.2.3.4", city: "Turin"},
			want: GeoInfo{City: "Turin", Country: "Italy"},
		},
		{
			name: "unknown 2-letter code is patched",
			old:  &GeoInfo{Country: "Zork Empire"},
			rat:  GeoRat{ip: "1.2.3.4", country: "ZO"},
			want: GeoInfo{Country: "Zork Empire"},
		},
		{
			name: "2-letter code is upgraded",
			old:  &GeoInfo{Country: "ZO"},
			rat:  GeoRat{ip: "1.2.3.4", country: "Zork Empire"},
			want: GeoInfo{Country: "Zork Empire"},
		},
		{
			name: "newest record wins on conflict",
			old:  &GeoInfo{City: "Turin", Country: "Italy"},
			rat:  GeoRat{ip: "1.2.3.4", city: "Paris", country: "France"},
			want: GeoInfo{City: "Paris", Country: "France"},
		},
		{
			name: "previous record without country",
			old:  &GeoInfo{},
			rat:  GeoRat{ip: "1.2.3.4", country: "Italy"},
			want: GeoInfo{Country: "Italy"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Squeal() panicked: %v", r)
				}
			}()
			ni := MakeNodeInfo()
			if tt.old != nil {
				ni.Geo.Set(tt.rat.ip, *tt.old)
			}
			tt.rat.Squeal(ni)
			if got := ni.Geo.Get(tt.rat.ip); got != tt.want {
				t.Errorf("Geo = %+v, want %+v", got, tt.want)
			}
			if !ni.EgressPoints.HasKey(tt.rat.ip) {
				t.Errorf("EgressPoints = %v, want %s", ni.EgressPoints.GetAllKeys(), tt.rat.ip)
			}
		})
	}
}

func TestGeoRatGeneric(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ip":
			fmt.Fprint(w, `{"ip": "1.2.3.4", "city": "Turin", "country": "IT"}`)
		case "/query":
			fmt.Fprint(w, `{"query": "5.6.7.8", "city": "Paris", "country": "France"}`)
		case "/garbage":
			fmt.Fprint(w, `<html>rate limited</html>`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	tests := []struct {
		path string
		want *GeoRat
	}{
		{"/ip", &GeoRat{source: srv.URL + "/ip", ip: "1.2.3.4", city: "Turin", country: "IT"}},
		{"/query", &GeoRat{source: srv.URL + "/query", ip: "5.6.7.8", city: "Paris", country: "France"}},
		{"/garbage", nil},
		{"/missing", nil},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			output := make(chan Rat, 1)
			geoRatGeneric(srv.URL+tt.path, context.Background(), output)
			close(output)
			got, ok := <-output
			switch {
			case tt.want == nil && ok:
				t.Errorf("got %+v, want nothing", got)
			case tt.want != nil && !ok:
				t.Errorf("got nothing, want %+v", *tt.want)
			case tt.want != nil && *got.(*GeoRat) != *tt.want:
				t.Errorf("got %+v, want %+v", *got.(*GeoRat), *tt.want)
			}
		})
	}
}

func TestQueryHostNameRats(t *testing.T) {
	host, err := os.Hostname()
	if err != nil {
		t.Skip("os.Hostname() failed:", err)
	}
	output := make(chan Rat, 1)
	QueryHostNameRats(output)
	if rat := (<-output).(*HostNameRat); rat.name != host {
		t.Errorf("name = %q, want %q", rat.name, host)
	}
}

func TestQueryIFaceRatsFindsLoopback(t *testing.T) {
	output := make(chan Rat, 256)
	QueryIFaceRats(output)
	close(output)
	for rat := range output {
		if rat.(*IFaceRat).loopback {
			return
		}
	}
	t.Error("no loopback interface reported")
}
