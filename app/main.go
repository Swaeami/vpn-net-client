package main

import (
	"context"
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
	INTERFACE_NAME = "swaeamiVPN"

	COORDINATOR_IP   = "127.0.0.1"
	COORDINATOR_PORT = 26100
	COORDINATOR_MTU  = 1024
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopChan := make(chan struct{}, 1)
	defer close(stopChan)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		gracefulShutdown := func() {
			log.Println("Gracefully stopping")
			cancel()
			go func() {
				time.Sleep(5 * time.Second)
				log.Println("Force stopping after 5 seconds")
				os.Exit(1)
			}()
		}
		select {
		case <-sigChan:
			gracefulShutdown()
		case <-stopChan:
			gracefulShutdown()
		}
	}()

	coordinator := coordinator.NewCoordinatorUDP(entities.CoordinatorConfig{
		IP:   COORDINATOR_IP,
		Port: COORDINATOR_PORT,
		MTU:  COORDINATOR_MTU,
	})
	tun := tun.NewTun(entities.TunConfig{
		InterfaceName: INTERFACE_NAME,
		IP:            THIS_IP,
		MTU:           MTU,
		Network:       NETWORK,
	})

	wg := sync.WaitGroup{}
	client := usecases.NewClient(coordinator, tun)

	wg.Add(1)
	go func() {
		defer wg.Done()
		client.Run(ctx, stopChan)
	}()
	log.Println("App started")

	wg.Wait()
	log.Println("App stopped")
}
