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
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.Config.Info.IP, c.Config.Info.Port))
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

func (c *CoordinatorUDP) Listen(ctx context.Context) {
	defer func() {
		c.Config.Wg.Done()
		c.Conn.Close()
	}()

	if c.Conn == nil {
		log.Printf("Error - udp Listen() before Connect()")
		return
	}

	buf := make([]byte, c.Config.Info.MTU)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Coordinator Listen")
			return
		default:
		}

		c.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		n, _, err := c.Conn.ReadFromUDP(buf[0:])
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Println("udp read error: ", err.Error())
			return
		}

		recieved := buf[:n]
		fmt.Println("> ", string(recieved))
		err = json.Unmarshal(recieved, &c.Config.VpnNet)
		if err != nil {
			log.Println("udp unmarshal error: ", err.Error())
			return
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
