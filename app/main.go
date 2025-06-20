package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
	"github.com/Swaeami/vpn-net/client/internal/repos/coordinator"
	"github.com/Swaeami/vpn-net/client/internal/repos/tun"
	"github.com/Swaeami/vpn-net/client/internal/usecases"
)

const (
	MTU            = 9000
	THIS_IP        = "26.10.0.10"
	NETWORK        = "26.10.0.0/24"
	INTERFACE_NAME = "utun151"

	COORDINATOR_IP   = "127.0.0.1"
	COORDINATOR_PORT = 26100
	COORDINATOR_MTU  = 1024
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var vpnNet entities.VpnNet
	wg := sync.WaitGroup{}
	wg.Add(2)

	coordinator := coordinator.NewCoordinatorUDP(entities.CoordinatorConfig{
		Info: entities.CoordinatorInfo{
			IP:   COORDINATOR_IP,
			Port: COORDINATOR_PORT,
			MTU:  COORDINATOR_MTU,
		},
		Wg:     &wg,
		VpnNet: &vpnNet,
	})
	tun := tun.NewTun(entities.TunConfig{
		Info: entities.TunInfo{
			InterfaceName: INTERFACE_NAME,
			IP:            THIS_IP,
			MTU:           MTU,
			Network:       NETWORK,
		},
		Wg:     &wg,
		VpnNet: &vpnNet,
	})

	err := tun.CheckAdminRights()
	if err != nil {
		log.Printf("admin rights error: %v", err)
		fmt.Scanln()
		return
	}

	client := usecases.NewClient(coordinator, tun)
	err = client.Connect(ctx)
	if err != nil {
		log.Println(err.Error())
		fmt.Scanln()
		return
	}
	log.Println("vpn started")

	go func() {
		<-sigChan
		log.Println("stopping...")
		cancel()
		// force stop
		go func() {
			time.Sleep(5 * time.Second)
			log.Println("gracefull stop did not succeed, force stop")
			os.Exit(1)
		}()
	}()

	wg.Wait()
	log.Println("vpn stopped")
}
