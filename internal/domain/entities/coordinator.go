package entities

import (
	"context"
	"sync"
)

type CoordinatorInfo struct {
	IP   string
	Port int
	MTU  int
}

type CoordinatorConfig struct {
	Info   CoordinatorInfo
	Wg     *sync.WaitGroup
	Ctx    *context.Context
	VpnNet *VpnNet
}
