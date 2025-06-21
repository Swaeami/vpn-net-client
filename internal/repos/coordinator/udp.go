package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Swaeami/vpn-net/client/internal/domain/entities"
)

type CoordinatorUDP struct {
	Config entities.CoordinatorConfig
	Conn   *net.UDPConn
}

func NewCoordinatorUDP(config entities.CoordinatorConfig) *CoordinatorUDP {
	return &CoordinatorUDP{Config: config}
}

func (c *CoordinatorUDP) Connect() error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.Config.IP, c.Config.Port))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}

	c.Conn = conn

	return nil
}

func (c *CoordinatorUDP) Listen(ctx context.Context, stopChan chan struct{}, vpnNet entities.VpnNet) {
	defer c.Conn.Close()

	if c.Conn == nil {
		log.Printf("Error - udp Listen() before Connect()")
		stopChan <- struct{}{}
		return
	}

	buf := make([]byte, c.Config.MTU)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Coordinator")
			return
		default:
		}

		c.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		n, _, err := c.Conn.ReadFromUDP(buf[0:])
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			log.Println("Coordinator read error: ", err.Error())
			stopChan <- struct{}{}
			continue
		}

		recieved := buf[:n]
		fmt.Println("> ", string(recieved))
		err = json.Unmarshal(recieved, &vpnNet)
		if err != nil {
			log.Println("udp unmarshal error: ", err.Error())
		}
	}

}

func (c *CoordinatorUDP) Send(data []byte) error {
	_, err := c.Conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}
