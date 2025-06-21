//go:build darwin

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
	config.Name = t.Config.Info.InterfaceName

	tunInterface, err := water.New(config)
	if err != nil {
		return err
	}

	cmd := exec.Command("ifconfig", t.Config.Info.InterfaceName, t.Config.Info.IP, t.Config.Info.IP, "up")
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("ifconfig", t.Config.Info.InterfaceName, "mtu", strconv.Itoa(t.Config.Info.MTU))
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("route", "add", "-net", t.Config.Info.Network, "-interface", t.Config.Info.InterfaceName)
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	t.Interface = tunInterface
	return nil
}

func (t *Tun) Destroy() error {
	if t.Interface == nil {
		return nil
	}

	cmd := exec.Command("route", "delete", "-net", t.Config.Info.Network)
	cmd.Output()

	t.Interface.Close()
	return nil
}
