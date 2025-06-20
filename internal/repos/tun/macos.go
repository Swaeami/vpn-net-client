//go:build darwin

package tun

import (
	"fmt"
	"os"
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

	cmd = exec.Command("r", t.Config.Info.InterfaceName, "mtu", strconv.Itoa(t.Config.Info.MTU))
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

func (t *Tun) CheckAdminRights() error {
	if os.Geteuid() != 0 {
		fmt.Println("requires admin rights")
		return t.restartWithSudo()
	}
	return nil
}

func (t *Tun) restartWithSudo() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path error: %w", err)
	}
	args := os.Args[1:]

	cmd := exec.Command("sudo", append([]string{executable}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("restart with sudo error: %w", err)
	}

	os.Exit(0)
	return nil
}
