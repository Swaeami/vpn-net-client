package tun

import (
	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
)

type TunBase struct {
	Config entities.TunConfig
}

func NewTunBase(config entities.TunConfig) *TunBase {
	return &TunBase{Config: config}
}

func (t *TunBase) GetConfig() entities.TunConfig {
	return t.Config
}
