//go:build darwin

package tun

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"

	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
	"github.com/songgao/water"
)

type Tun struct {
	TunBase
	Interface *water.Interface
}

func NewTun(config entities.TunConfig) *Tun {
	return &Tun{TunBase: *NewTunBase(config)}
}

func (t *Tun) Create() error {
	// we cant do fancy name on mac tun so we use any utun
	config := water.Config{
		DeviceType: water.TUN,
	}

	tunInterface, err := water.New(config)
	if err != nil {
		return err
	}
	interfaceName := tunInterface.Name()

	cmd := exec.Command("ifconfig", interfaceName, t.Config.IP, t.Config.IP, "up")
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("ifconfig", interfaceName, "mtu", strconv.Itoa(t.Config.MTU))
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("route", "add", "-net", t.Config.Network, "-interface", interfaceName)
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	t.Interface = tunInterface

	return nil
}

func (t *Tun) Read(ctx context.Context, stopChan chan struct{}, vpnNet entities.VpnNet) {

	if t.Interface == nil {
		log.Printf("Error - tun Read() before Create()")
		stopChan <- struct{}{}
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
			buf := make([]byte, t.Config.MTU)
			n, err := t.Interface.Read(buf)

			select {
			case readChan <- ReadResult{n, err, buf}:
			case <-ctx.Done():
				return
			}

			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping TUN by context")
			return
		case result, ok := <-readChan:
			if !ok {
				return
			}

			if result.err != nil {
				log.Printf("tun read error: %v", result.err)
				stopChan <- struct{}{}
				return
			}

			if result.n < 20 {
				log.Printf("too short packet: %d bytes", result.n)
				continue
			}

			destIP := net.IP(result.buf[16:20])
			found := false
			for _, ip := range vpnNet.IPs {
				fmt.Println(ip)
				if ip == destIP.String() {
					found = true
					break
				}
			}
			if found {
				log.Printf("Got %d bytes | IP from: %s | IP dest: %s\n", result.n, t.Config.IP, destIP.String())
			} else {
				log.Printf("Got %d bytes | IP from: %s to /dev/null", result.n, destIP.String())
			}
		}
	}
}

func (t *Tun) Destroy() error {
	if t.Interface == nil {
		return nil
	}

	cmd := exec.Command("route", "delete", "-net", t.Config.Network)
	cmd.Output()

	t.Interface.Close()
	return nil
}
