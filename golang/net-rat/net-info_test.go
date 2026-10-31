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
	"slices"
	"testing"
	"time"
)

type nodeSpec struct {
	host     string
	ifaces   []string
	privates []string
	egress   []string
	geo      map[string]string // ip -> country
	names    map[string]string // ip -> network name
}

func makeTestNodeInfo(s nodeSpec) *NodeInfo {
	ni := MakeNodeInfo()
	if s.host != "" {
		ni.HostName.Set(s.host, struct{}{})
	}
	for _, ip := range s.ifaces {
		ni.IFaceAddresses.SetZero(ip)
	}
	for _, ip := range s.privates {
		ni.PrivateAddresses.SetZero(ip)
	}
	for _, ip := range s.egress {
		ni.EgressPoints.SetZero(ip)
	}
	for ip, country := range s.geo {
		ni.Geo.Set(ip, GeoInfo{Country: country})
	}
	for ip, name := range s.names {
		ni.NetworkName.Set(ip, name)
	}
	return ni
}

func TestNodeInfoUpdate(t *testing.T) {
	tests := []struct {
		name     string
		spec     nodeSpec
		want     string
		complete bool
	}{
		{
			name: "no network",
			want: NO_NETWORK,
		},
		{
			name: "local only",
			spec: nodeSpec{host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}},
			want: "rat / 10.0.0.2 (local only)",
		},
		{
			name: "local only behind a tunnel",
			spec: nodeSpec{host: "rat", ifaces: []string{"192.168.1.5"}, privates: []string{"10.8.0.2"}},
			want: "rat / 192.168.1.5 -> 10.8.0.2 (local only)",
		},
		{
			name: "public without geo",
			spec: nodeSpec{host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}, egress: []string{"1.2.3.4"}},
			want: "rat / 10.0.0.2 => 1.2.3.4",
		},
		{
			name: "public with country",
			spec: nodeSpec{
				host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}, egress: []string{"1.2.3.4"},
				geo: map[string]string{"1.2.3.4": "Italy"},
			},
			want:     "rat / 10.0.0.2 => 1.2.3.4 (Italy)",
			complete: true,
		},
		{
			name: "public with country and network name",
			spec: nodeSpec{
				host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}, egress: []string{"1.2.3.4"},
				geo: map[string]string{"1.2.3.4": "Italy"}, names: map[string]string{"1.2.3.4": "GARR"},
			},
			want:     "rat / 10.0.0.2 => 1.2.3.4 (GARR, Italy)",
			complete: true,
		},
		{
			name: "tunnel with public",
			spec: nodeSpec{
				host: "rat", ifaces: []string{"192.168.1.5"}, privates: []string{"10.8.0.2"}, egress: []string{"1.2.3.4"},
				geo: map[string]string{"1.2.3.4": "Italy"},
			},
			want:     "rat / 192.168.1.5 -> 10.8.0.2 => 1.2.3.4 (Italy)",
			complete: true,
		},
		{
			name: "public address on the interface",
			spec: nodeSpec{
				host: "rat", ifaces: []string{"1.2.3.4"}, privates: []string{"1.2.3.4"}, egress: []string{"1.2.3.4"},
				geo: map[string]string{"1.2.3.4": "Italy"},
			},
			want:     "rat / 1.2.3.4 (Italy)",
			complete: true,
		},
		{
			name: "known CIDR overrides geo",
			spec: nodeSpec{
				host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}, egress: []string{"130.192.1.1"},
				geo: map[string]string{"130.192.1.1": "Italy"},
			},
			want:     "rat / 10.0.0.2 => 130.192.1.1 [Politecnico di Torino]",
			complete: true,
		},
		{
			name:     "known CIDR without geo",
			spec:     nodeSpec{host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}, egress: []string{"130.192.1.1"}},
			want:     "rat / 10.0.0.2 => 130.192.1.1 [Politecnico di Torino]",
			complete: true,
		},
		{
			name: "load-balanced egress",
			spec: nodeSpec{
				host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}, egress: []string{"1.2.3.4", "5.6.7.8"},
				geo: map[string]string{"1.2.3.4": "Italy", "5.6.7.8": "France"},
			},
			want:     "rat / 10.0.0.2 => " + MULTIPLE_IPS,
			complete: true,
		},
		{
			name: "no host name is never complete",
			spec: nodeSpec{
				ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}, egress: []string{"1.2.3.4"},
				geo: map[string]string{"1.2.3.4": "Italy"},
			},
			want: "10.0.0.2 => 1.2.3.4 (Italy)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ni := makeTestNodeInfo(tt.spec)
			complete := ni.Update()
			if ni.description != tt.want {
				t.Errorf("description = %q, want %q", ni.description, tt.want)
			}
			if complete != tt.complete {
				t.Errorf("Update() = %v, want %v", complete, tt.complete)
			}
		})
	}
}

func TestNodeInfoUpdatePicksIFaceFromLinkLayer(t *testing.T) {
	ni := makeTestNodeInfo(nodeSpec{
		host:     "rat",
		ifaces:   []string{"192.168.1.5", "172.17.0.1"},
		privates: []string{"10.8.0.2"},
	})
	ni.LinkLayerAddresses.Set("aa:bb:cc:dd:ee:ff", "192.168.1.5")
	ni.Update()
	if ni.iface != "192.168.1.5" {
		t.Errorf("iface = %q, want %q", ni.iface, "192.168.1.5")
	}
}

func TestNodeInfoUpdateIsIncremental(t *testing.T) {
	ni := makeTestNodeInfo(nodeSpec{host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}})
	if ni.Update() {
		t.Fatal("Update() complete without public IP")
	}
	ni.EgressPoints.SetZero("1.2.3.4")
	ni.Geo.Set("1.2.3.4", GeoInfo{Country: "Italy"})
	if !ni.Update() {
		t.Errorf("Update() incomplete after adding egress and geo: %q", ni.description)
	}
	if want := "rat / 10.0.0.2 => 1.2.3.4 (Italy)"; ni.description != want {
		t.Errorf("description = %q, want %q", ni.description, want)
	}
}

func TestNodeInfoCleanUp(t *testing.T) {
	now := time.Now()
	old := now.Add(-time.Hour)
	ancient := now.Add(-60 * 24 * time.Hour)

	ni := MakeNodeInfo()
	setAt(&ni.HostName, "rat", struct{}{}, old)
	setAt(&ni.IFaces, "aa:bb", "en0", old)
	setAt(&ni.IFaceAddresses, "10.0.0.2", struct{}{}, old)
	setAt(&ni.LinkLayerAddresses, "aa:bb", "10.0.0.2", old)
	setAt(&ni.PrivateAddresses, "10.0.0.2", struct{}{}, old)
	setAt(&ni.EgressPoints, "1.2.3.4", struct{}{}, old)
	setAt(&ni.EgressPoints, "5.6.7.8", struct{}{}, now)
	setAt(&ni.Geo, "1.2.3.4", GeoInfo{Country: "Italy"}, old)
	setAt(&ni.Geo, "9.9.9.9", GeoInfo{Country: "Switzerland"}, ancient)
	setAt(&ni.NetworkName, "1.2.3.4", "GARR", old)
	setAt(&ni.NetworkName, "9.9.9.9", "Quad9", ancient)

	ni.CleanUp(now.Add(-10*time.Minute), now.Add(-30*24*time.Hour))

	for name, n := range map[string]int{
		"HostName":           ni.HostName.Len(),
		"IFaces":             ni.IFaces.Len(),
		"IFaceAddresses":     ni.IFaceAddresses.Len(),
		"LinkLayerAddresses": ni.LinkLayerAddresses.Len(),
		"PrivateAddresses":   ni.PrivateAddresses.Len(),
	} {
		if n != 0 {
			t.Errorf("%s has %d entries, want 0", name, n)
		}
	}
	if !slices.Equal(ni.EgressPoints.GetAllKeys(), []string{"5.6.7.8"}) {
		t.Errorf("EgressPoints = %v, want [5.6.7.8]", ni.EgressPoints.GetAllKeys())
	}
	if !slices.Equal(ni.Geo.GetAllKeys(), []string{"1.2.3.4"}) {
		t.Errorf("Geo = %v, want [1.2.3.4]", ni.Geo.GetAllKeys())
	}
	if !slices.Equal(ni.NetworkName.GetAllKeys(), []string{"1.2.3.4"}) {
		t.Errorf("NetworkName = %v, want [1.2.3.4]", ni.NetworkName.GetAllKeys())
	}
}

func TestNodeInfoExtractIPs(t *testing.T) {
	ni := makeTestNodeInfo(nodeSpec{
		ifaces:   []string{"10.0.0.2"},
		privates: []string{"10.8.0.2"},
		egress:   []string{"1.2.3.4"},
		geo:      map[string]string{"1.2.3.4": "Italy", "5.6.7.8": "France"},
		names:    map[string]string{"9.9.9.9": "Quad9"},
	})
	ni.LinkLayerAddresses.Set("aa:bb", "10.0.0.2")

	got := ni.ExtractIPs()
	slices.Sort(got)
	want := []string{"1.2.3.4", "10.0.0.2", "5.6.7.8", "9.9.9.9"}
	if !slices.Equal(got, want) {
		t.Errorf("ExtractIPs() = %v, want %v", got, want)
	}
}

func TestDescribeNodeFromCache(t *testing.T) {
	useTempCache(t)
	ni := makeTestNodeInfo(nodeSpec{
		host: "rat", ifaces: []string{"10.0.0.2"}, privates: []string{"10.0.0.2"}, egress: []string{"1.2.3.4"},
		geo: map[string]string{"1.2.3.4": "Italy"},
	})
	if err := SaveCache(ni); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	got := DescribeNode(5 * time.Second)
	if want := "rat / 10.0.0.2 => 1.2.3.4 (Italy)"; got.description != want {
		t.Errorf("description = %q, want %q", got.description, want)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("DescribeNode took %v, expected an immediate answer from cache", elapsed)
	}
}

func TestUpdateCacheWithoutCache(t *testing.T) {
	useTempCache(t)
	if ni := UpdateCache(time.Second); ni != nil {
		t.Errorf("UpdateCache() = %v, want nil without a cache", ni)
	}
}

func TestUpdateCacheNothingToDo(t *testing.T) {
	useTempCache(t)
	ni := makeTestNodeInfo(nodeSpec{
		egress: []string{"1.2.3.4"},
		geo:    map[string]string{"1.2.3.4": "Italy"},
		names:  map[string]string{"1.2.3.4": "GARR"},
	})
	if err := SaveCache(ni); err != nil {
		t.Fatal(err)
	}
	if got := UpdateCache(time.Second); got != nil {
		t.Errorf("UpdateCache() = %v, want nil when geo and names are known", got)
	}
}
