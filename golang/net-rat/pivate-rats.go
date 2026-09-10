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
	"log/slog"
	"net"
	"sync"
	"time"
)

type PrivateRat struct {
	ip, source string
	loopback   bool
}

func (rat *PrivateRat) Squeal(ni *NodeInfo) {
	slog.Debug("PrivateRat squeals:", "src", rat.source, "ip", rat.ip, "loopback", rat.loopback)
	if !rat.loopback {
		ni.PrivateAddresses.SetZero(rat.ip)
		ni.Timestamp = time.Now()
	}
}

func QueryPrivateRats(ctx context.Context, output chan<- Rat) {
	var wg sync.WaitGroup
	wg.Add(1)
	go getIFaceAddr(&wg, output)
	wg.Add(len(ReliableIPs) * len(ReliablePorts))
	go getOutboundUdp(&wg, output)
	wg.Wait()
	output <- &TimeOutRat{}
}

func getIFaceAddr(wg *sync.WaitGroup, output chan<- Rat) {
	defer wg.Done()
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addrs {
			// Check the address type and exclude loopback (127.0.0.1)
			if ipnet, ok := address.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					output <- &PrivateRat{
						ip:       ipnet.IP.String(),
						source:   "InterfaceAddrs::ipnet",
						loopback: ipnet.IP.IsLoopback(),
					}
				}
			}
		}
	}
}

func getOutboundUdp(wg *sync.WaitGroup, output chan<- Rat) {
	for ip := range ReliableIPs {
		for p := range ReliablePorts {
			addr := net.JoinHostPort(ReliableIPs[ip], ReliablePorts[p])
			go func(udpAddress string, output chan<- Rat) {
				defer wg.Done()
				if conn, err := net.Dial("udp", udpAddress); err == nil {
					defer conn.Close()
					localAddr := conn.LocalAddr().(*net.UDPAddr)
					output <- &PrivateRat{
						ip:     localAddr.IP.String(),
						source: udpAddress,
					}
				} else {
					slog.Debug("Net failure:", "err", err.Error())
				}
			}(addr, output)
		}
	}
}
