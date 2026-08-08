//        ____()()     NetRat v0.3
//       /      @@     ~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"log"
	"log/slog"
	"net"
)

type HwRat struct {
	iface string
	ip    string
	hw    string
}

func (rat *HwRat) Squeal(ni *NodeInfo) {
	if rat.iface == "" {
		log.Panicf("HwRat squeaking with no interface!")
	}
	if rat.hw != "" && rat.ip != "" {
		slog.Debug("HwRat squeals:", "iface", rat.iface, "hw", rat.hw, "ip", rat.ip)
		ni.LinkLayerAddresses.Add(rat.hw)
		if _, found := ni.ResolutionTable[rat.ip]; !found {
			ni.ResolutionTable[rat.ip] = rat.hw
		} else {
			if ni.ResolutionTable[rat.ip] != rat.hw {
				log.Panicf("Conflicting bindings for %v (%v): %v vs. %v\n", rat.iface, rat.ip, ni.ResolutionTable[rat.ip], rat.hw)
			}
		}
		// clean up existing entries with the same iface
		for key, value := range ni.IFaces {
			if value == rat.iface {
				slog.Debug("Deleting old iface:", "k", rat.hw, "val", rat.iface)
				delete(ni.IFaces, key)
			}
		}
		ni.IFaces[rat.hw] = rat.iface
	}
}

func QueryHwRats(output chan<- Rat) {
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			//var ip net.IP
			switch addr.(type) {
			case *net.IPNet:
				output <- &HwRat{
					iface: iface.Name,
					ip:    addr.(*net.IPNet).IP.String(),
					hw:    iface.HardwareAddr.String(),
				}
			case *net.IPAddr:
				log.Fatalln("Panic: Unexpected *net.IPAddr")
			}
		}
	}
}
