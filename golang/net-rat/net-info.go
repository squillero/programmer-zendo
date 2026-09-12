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
	"slices"
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
	IFaceAddresses     Ephemeras[struct{}] `json:"iface_ips"`
	PrivateAddresses   Ephemeras[struct{}] `json:"private_ips"`
	LinkLayerAddresses Ephemeras[string]   `json:"link_layer_addresses"`
	EgressPoints       Ephemeras[struct{}] `json:"egress_points"`
	Geo                Ephemeras[GeoInfo]  `json:"geo_info"`
	NetworkName        Ephemeras[string]   `json:"net_name"`
	// NodeInfo readable description (not saved in json)
	host        string
	iface       string
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
		IFaceAddresses:     MakeEphemeras[struct{}](),
		LinkLayerAddresses: MakeEphemeras[string](),
		PrivateAddresses:   MakeEphemeras[struct{}](),
		EgressPoints:       MakeEphemeras[struct{}](),
		Geo:                MakeEphemeras[GeoInfo](),
		NetworkName:        MakeEphemeras[string](),
	}
}

func (ni *NodeInfo) CleanUp(cutOffTimeShort, cutOffTimeLong time.Time) {

	// Ephemeras
	tot := 0
	tot += ni.HostName.Invalidate(cutOffTimeShort)
	tot += ni.IFaces.Invalidate(cutOffTimeShort)
	tot += ni.IFaceAddresses.Invalidate(cutOffTimeShort)
	tot += ni.LinkLayerAddresses.Invalidate(cutOffTimeShort)
	tot += ni.PrivateAddresses.Invalidate(cutOffTimeShort)
	tot += ni.EgressPoints.Invalidate(cutOffTimeShort)
	tot += ni.Geo.Invalidate(cutOffTimeLong)
	tot += ni.NetworkName.Invalidate(cutOffTimeLong)

	if tot > 0 {
		slog.Info("Cleaned up invalid ephemeras:", "n", tot)
	} else {
		slog.Info("No invalid ephemeras")
	}
}

func (ni *NodeInfo) Update() bool {
	if newHost := ni.HostName.GetLatestKey(); newHost != "" {
		ni.host = newHost + " / "
	}

	var private string
	private = ni.PrivateAddresses.GetLatestKey()
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

	// iface
	var iface string
	if private != "" {
		switch ni.IFaceAddresses.Len() {
		case 0:
			iface = ""
		case 1:
			iface = ni.IFaceAddresses.GetLatestKey()
		default:
			ips := ni.LinkLayerAddresses.GetAllInfos()
			for _, i := range ni.IFaceAddresses.GetAllKeys() {
				if slices.Contains(ips, i) {
					iface = i
				}
			}
		}
		if iface != "" && iface != ni.iface {
			slog.Info("NodeInfo:", "iface", iface)
			ni.iface = iface
		}
	}

	var public string
	switch {
	case ni.EgressPoints.Len() == 0:
		public = ""
	case ni.EgressPoints.Len() == 1:
		public = ni.EgressPoints.GetLatestKey()
	case ni.EgressPoints.Len() > 1:
		// log.Println(ni.EgressPoints)
		public = MULTIPLE_IPS
	}
	if public != ni.public {
		slog.Info("NodeInfo:", "public", public)
		ni.public = public
	}

	if ni.public == MULTIPLE_IPS {
		// D'ho
		ni.geo = ""
	} else if ni.public != "" {
		if ni.Geo.HasKey(ni.public) && ni.NetworkName.Get(ni.public) != "" {
			ni.geo = fmt.Sprintf(" (%v, %v)", ni.NetworkName.Get(ni.public), ni.Geo.Get(ni.public).Country)
		} else if ni.Geo.HasKey(ni.public) {
			ni.geo = fmt.Sprintf(" (%v)", ni.Geo.Get(ni.public).Country)
		}
		for _, known := range KnownCIDR {
			if _, cidr, err := net.ParseCIDR(known.CIDR); err == nil {
				if cidr.Contains(net.ParseIP(ni.public)) {
					if _, used := usedCIDRs[known.CIDR]; !used {
						slog.Debug("KnownCIDR:", "cidr", cidr, "name", known.Name)
						usedCIDRs[known.CIDR] = struct{}{}
					}
					ni.geo = fmt.Sprintf(" [%v]", known.Name)
					break
				}
			} else {
				log.Panicln(err.Error())
			}
		}
	}

	var ip string
	switch {
	case ni.iface == "" && ni.private == "" && ni.public == "":
		ip = "No network detected"

	case ni.iface == ni.private && ni.private != "" && ni.public == "":
		ip = ni.private + " (local only)"
	case ni.iface != ni.private && ni.private != "" && ni.public == "":
		ip = ni.iface + " -> " + ni.private + " (local only)"

	case ni.iface == ni.private && ni.public != "" && ni.private == ni.public:
		ip = ni.public
	case ni.iface != ni.private && ni.public != "" && ni.private == ni.public:
		ip = ni.iface + " -> " + ni.public

	case ni.iface == ni.private && ni.private != "" && ni.public != "" && ni.private != ni.public:
		ip = ni.private + " => " + ni.public
	case ni.iface != ni.private && ni.private != "" && ni.public != "" && ni.private != ni.public:
		ip = ni.iface + " -> " + ni.private + " => " + ni.public
	}

	// return true if descr is completed
	ni.description = ni.host + ip + ni.geo
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
			cut = time.Now().Add(-10 * time.Minute)
		}
		ni.CleanUp(cut, time.Now().Add(-24*30*time.Hour))
	} else {
		ni = MakeNodeInfo()
	}
	if ni.Update() {
		slog.Info("All required info are in cache.")
		return ni
	}

	// I/O
	rats := make(chan Rat)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go QueryHostNameRats(rats)
	go QueryHwRats(rats)
	go QueryIFaceRats(rats)
	go QueryPrivateRats(ctx, rats)
	go QueryGlobalRats(ctx, rats)
	go QueryGeoRats(ctx, rats)
	go QueryRDAPRats(ctx, rats)

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
			slog.Info("NetworkInfo complete, delaying exit:", "εₜ", timeout/10)
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

	needGeo := false
	needRDAP := false
	for ip := range ips {
		if ni.Geo.HasKey(ip) {
			slog.Debug("UpdateCache::Known GeoInfo:", "ip", ip, "info", ni.Geo.Get(ip))
		} else {
			needGeo = true
		}
		if ni.NetworkName.Get(ip) == "" {
			go QueryRDAPRatsSingleIP(ctx, ip, rats)
			needRDAP = true
		} else {
			slog.Debug("UpdateCache::Known NetworkName:", "name", ni.NetworkName.Get(ip))
		}
	}
	if needGeo {
		go QueryGeoRats(ctx, rats)
	} else {
		slog.Info("UpdateCache: No GeoRats needed")
	}
	if !needRDAP {
		slog.Info("UpdateCache: No RDAPRats needed")
	}
	if needGeo || needRDAP {
		for {
			select {
			case lead := <-rats:
				lead.Squeal(ni)
			case <-ctx.Done():
				return ni
			}
		}
	}
	return nil
}
