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
	"log/slog"
)

var serviceGlobalURL = [...]string{
	// "http://api.ipify.org/",
	// "http://api4.ipify.org/",
	// "http://checkip.amazonaws.com/",
	"http://icanhazip.com/",
	"http://ident.me/",
	"http://ifconfig.me/ip",
	"http://ipecho.net/plain",
	// "https://checkip.amazonaws.com/",
	"https://icanhazip.com/",
}

type GlobalRat struct {
	ip, source string
}

func (rat *GlobalRat) Squeal(ni *NodeInfo) {
	slog.Debug("GlobalRat squeals:", "src", rat.source, "ip", rat.ip)
	ni.EgressPoints.Add(rat.ip)
}

func QueryGlobalRats(ctx context.Context, output chan<- Rat) {
	for i := range serviceGlobalURL {
		go func(url string) {
			if ip := FetchHttp(ctx, url); ip != "" {
				output <- &GlobalRat{
					ip:     ip,
					source: url,
				}
			}
		}(serviceGlobalURL[i])
	}
}
