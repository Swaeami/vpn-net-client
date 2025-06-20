package tun

import "github.com/songgao/water"

type TunInterface interface {
	Create() (*water.Interface, error)
}
