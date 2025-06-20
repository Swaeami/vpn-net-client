//go:build windows

package tun

import (
	"os/exec"
	"strconv"

	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
	"github.com/songgao/water"
)

type Tun struct {
	TunBase
}

func NewTun(config entities.TunConfig) *Tun {
	return &Tun{TunBase: *NewTunBase(config)}
}

func (t *Tun) Create() error {
	config := water.Config{
		DeviceType: water.TUN,
	}

	tunInterface, err := water.New(config)
	if err != nil {
		return err
	}

	cmd := exec.Command("netsh", "interface", "ip", "set", "address",
		"name="+t.Config.Info.InterfaceName, "static", t.Config.Info.IP, "255.255.255.0")
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("netsh", "interface", "ipv4", "set", "subinterface",
		t.Config.Info.InterfaceName, "mtu="+strconv.Itoa(t.Config.Info.MTU), "store=persistent")
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("route", "add", t.Config.Info.Network, "mask", "255.255.255.0", t.Config.Info.IP)
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	t.Interface = tunInterface

	return nil
}
