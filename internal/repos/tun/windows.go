//go:build windows

package tun

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

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

	time.Sleep(1 * time.Second)
	interfaceName := tunInterface.Name()

	powershellCmd := fmt.Sprintf(`New-NetIPAddress -InterfaceAlias "%s" -IPAddress %s -PrefixLength 24`, interfaceName, t.Config.Info.IP)

	cmd := exec.Command("powershell", "-Command", powershellCmd)
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("netsh", "interface", "ipv4", "set", "subinterface",
		interfaceName, "mtu="+strconv.Itoa(t.Config.Info.MTU), "store=persistent")
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	networkParts := strings.Split(t.Config.Info.Network, "/")

	cmd = exec.Command("route", "-p", "add", networkParts[0], "mask", "255.255.255.0", t.Config.Info.IP)
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

	cmd := exec.Command("route", "delete", t.Config.Info.Network, "mask", "255.255.255.0")
	cmd.Output()

	t.Interface.Close()
	return nil
}
