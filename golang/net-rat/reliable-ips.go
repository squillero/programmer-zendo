//        ____()()     NetRat v0.3
//       /      @@     ~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

var ReliableIPs = [...]string{
	// Alibaba Cloud
	"223.5.5.5",
	"223.5.5.5",
	//IPv6: "2400:3200::1",
	//IPv6: "2400:3200:baba::1",
	// Cloudflare
	"1.1.1.1",
	"1.0.0.1",
	//IPv6: "2606:4700:4700::1111",
	//IPv6: "2606:4700:4700::1001",
	// Google
	"8.8.8.8",
	"8.8.4.4",
	//IPv6: "2001:4860:4860::8888",
	//IPv6: "2001:4860:4860::8844",
	// OpenDNS (Cisco)
	"208.67.222.222",
	"208.67.220.220",
	//IPv6: "2620:119:35::35",
	//IPv6: "2620:119:53::53",
	// Quad9
	"9.9.9.9",
	"149.112.112.112",
	//IPv6: "2620:fe::fe",
	//IPv6: "2620:fe::9",
}

var ReliablePorts = [...]string{
	"43",  // WHOIS
	"53",  // DNS
	"80",  // HTTP
	"443", // HTTPS
}
