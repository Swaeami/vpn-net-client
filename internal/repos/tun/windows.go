//go:build windows

package tun

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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

func (t *Tun) CheckAdminRights() error {
	isAdmin, err := t.isRunningAsAdmin()
	if err != nil {
		return fmt.Errorf("admin rights error: %w", err)
	}

	if !isAdmin {
		fmt.Println("requires admin rights")
		return t.restartAsAdmin()
	}
	return nil
}

func (t *Tun) isRunningAsAdmin() (bool, error) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	isUserAnAdmin := shell32.NewProc("IsUserAnAdmin")

	ret, _, _ := isUserAnAdmin.Call()
	return ret != 0, nil
}

func (t *Tun) restartAsAdmin() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path error: %w", err)
	}

	args := os.Args[1:]
	argsStr := strings.Join(args, " ")

	psCmd := fmt.Sprintf(`Start-Process -FilePath "%s" -ArgumentList "%s" -Verb RunAs -Wait`, executable, argsStr)

	cmd := exec.Command("powershell", "-Command", psCmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("restart as admin error: %w", err)
	}

	os.Exit(0)
	return nil
}
