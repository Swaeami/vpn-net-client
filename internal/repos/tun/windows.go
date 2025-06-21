//go:build windows

package tun

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
	"github.com/Swaeami/vpn-net/client/pkg/winipcfg"
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/tun"
)

type Tun struct {
	TunBase
	Interface *tun.Device
}

func NewTun(config entities.TunConfig) *Tun {
	return &Tun{TunBase: *NewTunBase(config)}
}

func (t *Tun) Create() error {
	id := &windows.GUID{
		Data1: 0x0000000,
		Data2: 0xFFFF,
		Data3: 0xFFFF,
		Data4: [8]byte{0xFF, 0xe9, 0x76, 0xe5, 0x8c, 0x74, 0x06, 0x3e},
	}
	ifname := t.Config.InterfaceName
	ifce, err := tun.CreateTUNWithRequestedGUID(ifname, id, 0)
	if err != nil {
		return fmt.Errorf("unable to create TUN interface: %w", err)
	}

	nativeTunDevice := ifce.(*tun.NativeTun)
	link := winipcfg.LUID(nativeTunDevice.LUID())

	ip, err := netip.ParsePrefix(t.Config.IP + "/24")
	if err != nil {
		return fmt.Errorf("unable to set IP addresses: %w", err)
	}
	err = link.SetIPAddresses([]netip.Prefix{ip})
	if err != nil {
		return fmt.Errorf("unable to set IP addresses: %w", err)
	}

	cmd := exec.Command("netsh", "interface", "set", "interface", t.Config.InterfaceName, "admin=enable")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("unable to enable interface: %s\n", string(output))
	}

	link.AddRoute(netip.PrefixFrom(netip.MustParseAddr(t.Config.IP), 24), netip.MustParseAddr(t.Config.IP), 1)

	cmd = exec.Command("netsh", "interface", "ipv4", "set", "subinterface",
		t.Config.InterfaceName, "mtu="+strconv.Itoa(t.Config.MTU), "store=persistent")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Предупреждение: не удалось установить MTU: %s\n", string(output))
	}

	t.Interface = &ifce

	return nil
}

func (t *Tun) Destroy() error {
	if t.Interface == nil {
		return nil
	}

	cmd := exec.Command("route", "delete", t.Config.Network, "mask", "255.255.255.0")
	cmd.Output()

	(*t.Interface).Close()
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
			buf := make([][]byte, 1)
			buf[0] = make([]byte, t.Config.MTU)
			sizes := []int{t.Config.MTU}
			_, err := (*t.Interface).Read(buf, sizes, 0)

			select {
			case readChan <- ReadResult{sizes[0], err, buf[0]}:
			case <-ctx.Done():
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

			// походу на винде надо отдельно фильтровать пакеты хз
			destIP := net.IP(result.buf[16:20])
			_, subnet, _ := net.ParseCIDR("26.10.0.0/24")
			if !subnet.Contains(destIP) {
				continue
			}
			destIPsplit := strings.Split(destIP.String(), ".")
			if destIPsplit[3] == "0" || destIPsplit[3] == "255" {
				continue
			}

			if result.n < 20 {
				log.Printf("too short packet: %d bytes", result.n)
				log.Printf("packet: %v", result.buf[:result.n])
				continue
			}

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
