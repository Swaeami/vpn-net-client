package ports

import (
	"context"

	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
)

type TunManager interface {
	Create() error
	Read(ctx context.Context)
	GetConfig() entities.TunConfig
	Destroy() error
}
