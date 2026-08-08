//        ____()()     NetRat v0.3
//       /      @@     ~~~~~~~~~~~
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
	"log"
	"log/slog"
	"strings"
	"time"
)

type GeoRat struct {
	source        string
	ip            string
	city, country string
}

// Country represents the ISO 3166 structure
type Country struct {
	Name    string `json:"name"`
	Alpha2  string `json:"alpha2"`
	Alpha3  string `json:"alpha3"`
	Numeric string `json:"numeric"`
}

var Countries []Country

//go:embed data/iso3166.json
var countriesJSON []byte

// Runs automatically BEFORE main()
func init() {
	// Parse the embedded JSON once at startup
	if err := json.Unmarshal(countriesJSON, &Countries); err != nil {
		log.Fatalf("Critical startup error: failed to parse country data: %v", err)
	}
}

func (rat *GeoRat) Squeal(ni *NodeInfo) {
	rat.country = canonize(rat.country)
	slog.Debug("GeoRat squeals:", "src", rat.source, "ip", rat.ip, "city", rat.city, "country", rat.country)
	if val, found := ni.Geo[rat.ip]; found {
		// Found it!
		switch {
		case val.City != "" && rat.city == "":
			slog.Debug("Patching:", "city", val.City)
		case val.City != "" && rat.city != "" && val.City != rat.city:
			slog.Debug("Inconsistent record:", "old", val, "new", *rat)
		}
		switch {
		case val.Country != "" && rat.country == "":
			slog.Debug("Patching:", "country", val.Country)
			rat.country = val.Country
		case strings.EqualFold(val.Country[:2], rat.country[:2]) && len(rat.country) == 2:
			slog.Debug("Patching:", "country", val.Country)
			rat.country = val.Country
		case len(val.Country) == 2 && len(rat.country) > 2:
			slog.Debug("Updating:", "country", rat.country)
		case val.Country != "" && rat.country != "" && val.Country != rat.country:
			slog.Warn("Inconsistent record:", "old", val, "new", *rat)
		}
	}
	ni.Geo[rat.ip] = GeoInfo{Timestamp: time.Now(), City: rat.city, Country: rat.country}
	ni.EgressPoints.Add(rat.ip)
}

func QueryGeoRats(ctx context.Context, output chan<- Rat) {
	time.Sleep(250 * time.Microsecond) // Let's wait for a fast public rat...
	go geoRatAirVPN(ctx, output)
	go geoRatGeneric("https://am.i.mullvad.net/json", ctx, output)
	go geoRatGeneric("http://ipinfo.io/json", ctx, output)
	go geoRatGeneric("http://ip-api.com/json", ctx, output)
	go geoRatGeneric("http://ipwho.is/", ctx, output)

}

func geoRatAirVPN(ctx context.Context, output chan<- Rat) {
	SOURCE := "https://airvpn.org/api/whatismyip/"
	type airvpn struct {
		Ip  string `json:"ip"`
		Geo struct {
			Country string `json:"name"`
		} `json:"geo"`
		GeoAdditional struct {
			City string `json:"city_name"`
		} `json:"geo_additional"`
	}

	var info airvpn
	if raw := FetchHttp(ctx, SOURCE); raw != "" {
		if json.Unmarshal([]byte(raw), &info) == nil {
			output <- &GeoRat{
				ip:      info.Ip,
				source:  SOURCE,
				city:    info.GeoAdditional.City,
				country: info.Geo.Country,
			}
		}
	}
}

func geoRatGeneric(url string, ctx context.Context, output chan<- Rat) {
	type generic struct {
		Ip1     string `json:"ip"`
		Ip2     string `json:"query"`
		City    string `json:"city"`
		Country string `json:"country"`
	}

	var info generic
	if raw := FetchHttp(ctx, url); raw != "" {
		if json.Unmarshal([]byte(raw), &info) == nil {
			if info.Ip1 != "" && info.Ip2 != "" {
				log.Panicln("Found both json:ip and json:query!")
			}
			output <- &GeoRat{
				ip:      info.Ip1 + info.Ip2,
				source:  url,
				city:    info.City,
				country: info.Country,
			}
		}
	}
}

func canonize(geo string) string {
	g := strings.ReplaceAll(strings.ToUpper(geo), ".", "")
	for _, iso3166 := range Countries {
		if g == iso3166.Alpha2 || g == iso3166.Alpha3 {
			slog.Debug("canonize:", "old", geo, "new", iso3166.Name)
			return iso3166.Name
		}
	}
	return geo
}
