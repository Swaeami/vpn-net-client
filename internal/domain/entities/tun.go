package entities

import (
	"context"
	"sync"
)

type TunInfo struct {
	InterfaceName string
	IP            string
	MTU           int
	Network       string
}

type TunConfig struct {
	Info   TunInfo
	Wg     *sync.WaitGroup
	Ctx    *context.Context
	VpnNet *VpnNet
}
