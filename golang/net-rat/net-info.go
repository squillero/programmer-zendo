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

const MULTIPLE_IPS = "multiple IPs (load-balanced egress)"

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
	City    string `json:"city"`    // Physical location of the egress gateway
	Country string `json:"country"` // Physical location of the egress gateway
}

type NodeInfoStatus int

const (
	NIS_STILL_WORKING NodeInfoStatus = 0
	NIS_VALID         NodeInfoStatus = 1
	NIS_ABORTED       NodeInfoStatus = -1
	LOOPBACK          string         = "127.0.0.1"
	NO_NETWORK        string         = "No network detected"
)

type NName struct {
	Ip          string
	NetworkName string
}

// Comprehensive description of a node: MACs, local IPs, egress points, ...
type NodeInfo struct {
	// NodeInfo internal state
	SchemaVersion      string              `json:"schema"`
	Timestamp          time.Time           `json:"timestamp"`
	HostName           Ephemeras[struct{}] `json:"hostname"`
	IFaces             Ephemeras[string]   `json:"ifaces"`
	LinkLayerAddresses Ephemeras[string]   `json:"link_layer_addresses"`
	PrivateAddresses   Ephemeras[struct{}] `json:"private_ips"`
	EgressPoints       Ephemeras[struct{}] `json:"egress_points"`
	// ResolutionTable    Ephemeras[string]   `json:"bindings"`
	Geo         Ephemeras[GeoInfo] `json:"geo_info"`
	NetworkName Ephemeras[string]  `json:"net_name"`
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
		SchemaVersion:      RAT_VERSION,
		Timestamp:          time.Now(),
		HostName:           MakeEphemeras[struct{}](),
		IFaces:             MakeEphemeras[string](),
		LinkLayerAddresses: MakeEphemeras[string](),
		PrivateAddresses:   MakeEphemeras[struct{}](),
		EgressPoints:       MakeEphemeras[struct{}](),
		//ResolutionTable:    MakeEphemeras[string](),
		Geo:         MakeEphemeras[GeoInfo](),
		NetworkName: MakeEphemeras[string](),
	}
}

func (ni *NodeInfo) CleanUp(cutOffTime time.Time) {

	// Ephemeras
	tot := 0
	tot += ni.PrivateAddresses.Invalidate(time.Now())
	tot += ni.HostName.Invalidate(time.Now())

	tot += ni.LinkLayerAddresses.Invalidate(cutOffTime)
	tot += ni.EgressPoints.Invalidate(cutOffTime)
	//tot += ni.ResolutionTable.Invalidate(cutOffTime)
	tot += ni.Geo.Invalidate(cutOffTime)
	if tot > 0 {
		slog.Info("Cleaned up invalid ephemeras:", "n", tot)
	} else {
		slog.Info("No invalid ephemeras")
	}
}

func (ni *NodeInfo) Update() bool {
	if newHost := ni.HostName.LatestKey(); newHost != "" {
		if newHost != "" && newHost != ni.host {
			slog.Info("NodeInfo:", "host", ni.HostName)
		}
		ni.host = newHost + " / "
	}

	var private string
	private = ni.PrivateAddresses.LatestKey()
	switch {
	case private == "":
		private = LOOPBACK // default
	case ni.PrivateAddresses.Len() > 1:
		slog.Debug("Update", "Multiple PrivateAddresses", ni.PrivateAddresses.TimedInfo)
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
		public = ni.EgressPoints.LatestKey()
	case ni.EgressPoints.Len() > 1:
		// log.Println(ni.EgressPoints)
		public = MULTIPLE_IPS
	}
	if public != ni.public {
		slog.Info("NodeInfo:", "public", public)
		ni.public = public
	}

	oldGeo := ni.geo
	geoStr := ""
	if ni.public != "" {
		if ni.Geo.HasKey(ni.public) {
			ni.geo = ni.Geo.Get(ni.public).Country
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

	// D'ho
	ni.net = ni.NetworkName.Get(ip)

	// return true if descr is completed
	ni.description = ni.host + ip + geoStr
	return (ni.host != "" && ni.private != "" && ni.public != "" && ni.geo != "") ||
		(ni.host != "" && ni.private != "" && ni.public == MULTIPLE_IPS)
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

func DescribeNode(timeout time.Duration) *NodeInfo {
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
		if ni.Update() && !alreadyHolding {
			slog.Error("NetworkInfo complete, delaying exit:", "εₜ", timeout/10)
			go func() {
				time.Sleep(timeout / 10)
				cancel() // cancel() is designed to be thread-safe and idempotent
			}()
			alreadyHolding = true
		}
	}
	return ni
}

func UpdateCache(timeout time.Duration) *NodeInfo {
	var ni *NodeInfo
	if ni = LoadCache(); ni == nil {
		return nil
	}

	// I/O
	rats := make(chan Rat)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ips := make(map[string]struct{}, 0)
	for ip, _ := range ni.Geo.Content {
		ips[ip] = struct{}{}
	}
	for ip, _ := range ni.EgressPoints.Content {
		ips[ip] = struct{}{}
	}

	go QueryGeoRats(ctx, rats)
	for ip := range ips {
		slog.Error("RDAP", "ip", ip)
		go QueryRDAPRats(ctx, ip, rats)
	}

	for {
		select {
		case lead := <-rats:
			lead.Squeal(ni)
		case <-ctx.Done():
			return ni
		}
	}
}
