package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Swaeami/vpn-net/client/helpers/admin"
	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
	"github.com/Swaeami/vpn-net/client/internal/domain/ports"
)

type Client struct {
	coordinator ports.NetworkCoordinator
	tunManager  ports.TunManager
}

func NewClient(coordinator ports.NetworkCoordinator, tunManager ports.TunManager) *Client {
	return &Client{coordinator: coordinator, tunManager: tunManager}
}

func (c *Client) Run(ctx context.Context, stopChan chan struct{}) error {
	adminRights, err := admin.IsAdmin()
	if err != nil {
		return err
	}
	if !adminRights {
		return fmt.Errorf("admin rights not granted")
	}

	err = c.tunManager.Create()
	if err != nil {
		return err
	}

	err = c.coordinator.Connect()
	if err != nil {
		return err
	}

	vpnNet := entities.VpnNet{}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.tunManager.Read(ctx, stopChan, vpnNet)
	}()
	go func() {
		defer wg.Done()
		c.coordinator.Listen(ctx, stopChan, vpnNet)
	}()

	netRequest := entities.NetRequest{
		TunIP: c.tunManager.GetConfig().IP,
		Type:  "add",
	}
	netRequestBytes, err := json.Marshal(netRequest)
	if err != nil {
		return err
	}
	err = c.coordinator.Send(netRequestBytes)
	if err != nil {
		return err
	}

	wg.Wait()

	return nil
}
