package usecases

import (
	"context"
	"encoding/json"
	"fmt"

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

func (c *Client) Connect(ctx context.Context) error {
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

	go c.tunManager.Read(ctx)
	go c.coordinator.Listen(ctx)

	netRequest := entities.NetRequest{
		TunIP: c.tunManager.GetConfig().Info.IP,
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

	return nil
}
