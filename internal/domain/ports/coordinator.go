package ports

import "context"

type NetworkCoordinator interface {
	Connect() error
	Listen(ctx context.Context)
	Send(data []byte) error
}
