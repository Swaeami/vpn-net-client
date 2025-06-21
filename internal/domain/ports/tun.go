package ports

import (
	"context"

	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
)

type TunManager interface {
	Create() error
	Read(ctx context.Context, stopChan chan struct{}, vpnNet entities.VpnNet)
	GetConfig() entities.TunConfig
	Destroy() error
}
