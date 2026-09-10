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
	"encoding/json"
	"log/slog"
	"time"
)

const RDAP_SERVER = "https://rdap.org/"

// Minimal struct to parse relevant domain fields
type RDAPDomain struct {
	StartAddress string `json:"startAddress"`
	EndAddress   string `json:"endAddress"`
	Name         string `json:"name"`
	Country      string `json:"country"`
}

type RDAPRat struct {
	ip      string
	name    string
	country string
}

func (rat *RDAPRat) Squeal(ni *NodeInfo) {
	slog.Info("Pre Squeal", "ni", ni.NetworkName)
	rat.country = canonize(rat.country)
	slog.Info("RDAPRat squeals:", "name", rat.name, "country", rat.country)
	slog.Error("RDAPRat squeals:", "name", rat.name, "country", rat.country)
	ni.NetworkName.Set(rat.ip, rat.name)
	slog.Info("Post Squeal", "ni", ni.NetworkName)
	ni.Timestamp = time.Now()
}

func QueryRDAPRats(ctx context.Context, ip string, output chan<- Rat) {
	slog.Warn("RDAP!", "ip", ip)
	src := RDAP_SERVER + "/ip/" + ip
	raw := FetchRaw(map[string]string{"Accept": "application/rdap+json"}, ctx, src)
	if raw == nil {
		return
	}
	var info RDAPDomain
	if err := json.Unmarshal(raw, &info); err != nil {
		slog.Error("QueryRDAPRats", "Unmarshal", err)
	}
	slog.Info("RDAP", "rat", info)
	output <- &RDAPRat{
		ip:      ip,
		name:    info.Name,
		country: info.Country,
	}
	output <- &GeoRat{
		ip:      ip,
		source:  RDAP_SERVER,
		city:    "",
		country: info.Country,
	}
}
