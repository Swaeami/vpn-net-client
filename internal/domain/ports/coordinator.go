package ports

import (
	"context"

	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
)

type NetworkCoordinator interface {
	Connect() error
	Listen(ctx context.Context, stopChan chan struct{}, vpnNet entities.VpnNet)
	Send(data []byte) error
}
