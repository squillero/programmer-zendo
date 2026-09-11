//        ____()()     NetRat v0.2.4
//       /      @@     ~~~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"log/slog"
	"net"
	"time"
)

type IFaceRat struct {
	ip       string
	loopback bool
}

func (rat *IFaceRat) Squeal(ni *NodeInfo) {
	slog.Debug("IFaceRat squeals:", "ip", rat.ip, "loopback", rat.loopback)
	if !rat.loopback {
		ni.IFaceAddresses.SetZero(rat.ip)
		ni.Timestamp = time.Now()
	}
}

func QueryIFaceRats(output chan<- Rat) {
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addrs {
			// Check the address type and exclude loopback (127.0.0.1)
			if ipnet, ok := address.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					output <- &IFaceRat{
						ip:       ipnet.IP.String(),
						loopback: ipnet.IP.IsLoopback(),
					}
				}
			}
		}
	}
}
