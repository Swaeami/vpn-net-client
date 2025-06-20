package tun

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
	"github.com/songgao/water"
)

type TunBase struct {
	Config    entities.TunConfig
	Interface *water.Interface
}

func NewTunBase(config entities.TunConfig) *TunBase {
	return &TunBase{Config: config}
}

func (t *TunBase) Read(ctx context.Context) {
	defer t.Config.Wg.Done()

	if t.Interface == nil {
		log.Printf("Error - tun Read() before Create()")
		return
	}

	type ReadResult struct {
		n   int
		err error
		buf []byte
	}

	readChan := make(chan ReadResult, 1)

	go func() {
		defer close(readChan)
		for {
			buf := make([]byte, t.Config.Info.MTU)
			n, err := t.Interface.Read(buf)

			select {
			case readChan <- ReadResult{n, err, buf}:
			case <-ctx.Done():
				return
			}

			if err != nil {
				// При ошибке завершаем горутину
				return
			}
		}
	}()

	// Основной цикл обработки
	for {
		select {
		case <-ctx.Done():
			log.Println("stopping tun")
			return
		case result, ok := <-readChan:
			if !ok {
				return
			}

			if result.err != nil {
				log.Printf("tun read error: %v", result.err)
				return
			}

			if result.n < 20 {
				log.Printf("too short packet: %d bytes", result.n)
				continue
			}

			destIP := net.IP(result.buf[16:20])
			found := false
			for _, ip := range t.Config.VpnNet.IPs {
				fmt.Println(ip)
				if ip == destIP.String() {
					found = true
					break
				}
			}
			if found {
				log.Printf("Got %d bytes | IP from: %s | IP dest: %s\n", result.n, t.Config.Info.IP, destIP.String())
			} else {
				log.Printf("Got %d bytes | IP from: %s to /dev/null", result.n, destIP.String())
			}
		}
	}
}

func (t *TunBase) GetConfig() entities.TunConfig {
	return t.Config
}
