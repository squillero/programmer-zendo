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
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"
)

// Country represents the ISO 3166 structure
type knownCIDR struct {
	CIDR string `json:"cidr"`
	Name string `json:"name"`
}

var KnownCIDR []knownCIDR

//go:embed data/CIDRs.json
var knownCIDRsJSON []byte

// Runs automatically BEFORE main()
func init() {
	// Parse the embedded JSON once at startup
	if err := json.Unmarshal(knownCIDRsJSON, &KnownCIDR); err != nil {
		log.Fatalf("Critical startup error: failed to parse known CIDRs data: %v", err)
	}
}

var usedCIDRs = make(map[string]struct{})

// Gateway
type GeoInfo struct {
	Timestamp time.Time `json:"timestamp"`
	City      string    `json:"city"`    // Physical location of the egress gateway
	Country   string    `json:"country"` // Physical location of the egress gateway
}

type NodeInfoStatus int

const (
	NIS_STILL_WORKING NodeInfoStatus = 0
	NIS_VALID         NodeInfoStatus = 1
	NIS_ABORTED       NodeInfoStatus = -1
	LOOPBACK          string         = "127.0.0.1"
	NO_NETWORK        string         = "No network detected"
)

// Comprehensive description of a node: MACs, local IPs, egress points, ...
type NodeInfo struct {
	// NodeInfo internal state
	SchemaVersion      string             `json:"schema"`
	Timestamp          time.Time          `json:"timestamp"`
	HostName           string             `json:"hostname"`
	IFaces             map[string]string  `json:"ifaces"`
	LinkLayerAddresses Ephemeras          `json:"link_layer_addresses"`
	PrivateAddresses   Ephemeras          `json:"private_ips"`
	EgressPoints       Ephemeras          `json:"egress_points"`
	ResolutionTable    map[string]string  `json:"bindings"`
	Geo                map[string]GeoInfo `json:"geo_info"`
	NetworkName        map[string]string  `json:"net_name"`
	// NodeInfo readable description (not saved in json)
	host        string
	private     string
	vpn         string
	public      string
	geo         string
	net         string
	description string
}

func MakeNodeInfo() *NodeInfo {
	return &NodeInfo{
		SchemaVersion:   RAT_VERSION,
		Timestamp:       time.Now(),
		ResolutionTable: make(map[string]string),
		IFaces:          make(map[string]string),
		Geo:             make(map[string]GeoInfo),
		NetworkName:     make(map[string]string)}
}

func (ni *NodeInfo) CleanUp(cutTime time.Time) {

	// Ephemeras
	tot := 0
	tot += ni.LinkLayerAddresses.Invalidate(cutTime)
	tot += ni.PrivateAddresses.Invalidate(cutTime)
	tot += ni.EgressPoints.Invalidate(cutTime)
	if tot > 0 {
		slog.Info("Cleaned up invalid ephemeras:", "n", tot)
	} else {
		slog.Info("No invalid ephemeras")
	}

	// Cleanup useless mappings
	tot = 0
	for key := range ni.ResolutionTable {
		if ni.LinkLayerAddresses.Find(key).IsZero() && ni.PrivateAddresses.Find(key).IsZero() && ni.EgressPoints.Find(key).IsZero() {
			slog.Debug("Deleting unused mapping:", "k", key, "val", ni.ResolutionTable[key])
			delete(ni.ResolutionTable, key)
			tot++
		} else {
			slog.Debug("Found valid mapping:", "k", key, "val", ni.ResolutionTable[key])
		}
	}
	if tot > 0 {
		slog.Info("Cleaned up unused mappings:", "n", tot)
	} else {
		slog.Info("No unused mappings")
	}

	// Geo Info
	tot = 0
	geoCutOff := time.Now().Add(-10 * 24 * time.Hour)
	for k, g := range ni.Geo {
		if !g.Timestamp.After(geoCutOff) {
			slog.Debug("Removing old GeoInfo:", "k", k, "val", g.Country, "ΔT", time.Since(g.Timestamp))
			delete(ni.Geo, k)
			tot++
		} else {
			slog.Debug("Found valid GeoInfo:", "k", k, "val", g.Country, "ΔT", time.Since(g.Timestamp))
		}
	}
	if tot > 0 {
		slog.Info("Cleaned up outdated geo info:", "n", tot)
	} else {
		slog.Info("No outdated geo info")
	}
}

func (ni *NodeInfo) Update() bool {
	oldHost := ni.host
	ni.host = ni.HostName + " / "
	if ni.HostName != "" && ni.host != oldHost {
		slog.Info("NodeInfo:", "host", ni.HostName)
	}

	var private string
	switch {
	case ni.PrivateAddresses.Len() == 0:
		private = LOOPBACK // default
	case ni.PrivateAddresses.Len() == 1:
		private = ni.PrivateAddresses.Peek()
	case ni.PrivateAddresses.Len() > 1:
		slog.Warn("Update", "PrivateAddresses", ni.PrivateAddresses)
		for _, ip := range ni.PrivateAddresses.Content {
			if _, found := ni.ResolutionTable[ip]; found {
				private = ip
				break
			}
		}
	}
	if private != "" && private != LOOPBACK && private != ni.private {
		slog.Info("NodeInfo:", "private", private)
		ni.private = private
	}

	var public string
	switch {
	case ni.EgressPoints.Len() == 0:
		public = ""
	case ni.EgressPoints.Len() == 1:
		public = ni.EgressPoints.Peek()
	case ni.EgressPoints.Len() > 1:
		// log.Println(ni.EgressPoints)
		public = "multiple IPs (load-balanced egress)"
	}
	if public != ni.public {
		slog.Info("NodeInfo:", "public", public)
		ni.public = public
	}

	oldGeo := ni.geo
	geoStr := ""
	if ni.public != "" {
		if g, ok := ni.Geo[ni.public]; ok {
			ni.geo = g.Country
			geoStr = fmt.Sprintf(" (%v)", ni.geo)
		}
		for _, known := range KnownCIDR {
			if _, cidr, err := net.ParseCIDR(known.CIDR); err == nil {
				if cidr.Contains(net.ParseIP(ni.public)) {
					ni.geo = known.Name
					if _, used := usedCIDRs[known.CIDR]; !used {
						slog.Debug("KnownCIDR:", "cidr", cidr, "name", known.Name)
						usedCIDRs[known.CIDR] = struct{}{}
					}
					geoStr = fmt.Sprintf(" [%v]", ni.geo)
					break
				}
			} else {
				log.Panicln(err.Error())
			}
		}
	}
	if ni.geo != oldGeo {
		slog.Info("NodeInfo:", "geo", ni.geo)
	}

	var ip string
	switch {
	case ni.private == "" && ni.public != "":
		log.Panicln("Panic: No network detected, but valid egress info")
	case ni.private == "" && ni.public == "":
		ip = "No network detected"
	case ni.private != "" && ni.public == "":
		ip = ni.private + " (local only)"
	case ni.private != "" && ni.public != "" && ni.private == ni.public:
		ip = ni.public
	case ni.private != "" && ni.public != "" && ni.private != ni.public:
		ip = ni.private + " => " + ni.public
	}

	if name, ok := ni.NetworkName[ip]; ok {
		ni.net = name
	} else {
		ni.net = ""
	}

	ni.description = ni.host + ip + geoStr
	// return true if descr is completed
	return ni.host != "" && ni.private != "" && ni.public != "" && ni.geo != ""
}

// A generic Rat can update the node info
type Rat interface {
	Squeal(ni *NodeInfo)
}

// TimeOut Rats
type TimeOutRat struct {
}

func (TimeOutRat) Squeal(ni *NodeInfo) {
	log.Panicln("TimeOut Rats do not squeal!")
}

func DescribeNode(timeout, timeoutExt time.Duration) *NodeInfo {
	var ni *NodeInfo
	if ni = LoadCache(); ni != nil {
		var cut time.Time
		if InvalidateEphemeras {
			cut = time.Now()
		} else {
			cut = time.Now().Add(-666 * time.Minute)
		}
		ni.CleanUp(cut)
	} else {
		ni = MakeNodeInfo()
	}

	// I/O
	rats := make(chan Rat)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go QueryHostNameRats(rats)
	go QueryHwRats(rats)
	go QueryPrivateRats(ctx, rats)
	go QueryGlobalRats(ctx, rats)
	go QueryGeoRats(ctx, rats)

	alreadyHolding := false
	startTime := time.Now()
	active := !ni.Update()
	for active {
		select {
		case lead := <-rats:
			if _, timedOut := lead.(*TimeOutRat); timedOut {
				if ni.private == LOOPBACK {
					slog.Debug("Can't connect, giving up.", "ΔT", time.Since(startTime))
					cancel()
				}
			} else {
				lead.Squeal(ni)
			}
		case <-ctx.Done():
			slog.Debug("Context \"done\" signal received.")
			active = false
			continue
		}
		if ni.Update() && !alreadyHolding && timeoutExt >= 0 {
			slog.Debug("NetworkInfo complete, delaying exit:", "εₜ", timeoutExt)
			go func() {
				time.Sleep(timeoutExt)
				cancel() // cancel() is designed to be thread-safe and idempotent
			}()
			alreadyHolding = true
		}
	}
	return ni
}
