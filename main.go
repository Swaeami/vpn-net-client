package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/Swaeami/vpn-net-grpc/client-macos/internal/infrastructure/tun"
	"github.com/songgao/water"
)

const (
	MTU            = 9000
	THIS_IP        = "26.10.0.10"
	NETWORK        = "26.10.0.0/24"
	INTERFACE_NAME = "utun151"

	SIGNALIZER_HOST = "127.0.0.1:26100"
	SIGNALIZER_MTU  = 1024
)

type VpnNet struct {
	Name string
	IPs  []string
}

type NetRequest struct {
	TunIP string
	Type  string
}

func CreateTun() (*water.Interface, error) {
	config := tun.TunConfig{
		InterfaceName: INTERFACE_NAME,
		IP:            THIS_IP,
		MTU:           MTU,
		Network:       NETWORK,
	}

	tunInterface := tun.NewTunManager(config)
	return tunInterface.Create()
}

func ConnectToSingalizer() (*net.UDPConn, error) {
	updAddr, err := net.ResolveUDPAddr("udp", SIGNALIZER_HOST)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, updAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func ReadTun(tun *water.Interface, vpnNet *VpnNet) {
	buf := make([]byte, MTU)
	for {
		n, err := tun.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if n < 20 {
			log.Printf("Packet too short: %d bytes", n)
			continue
		}
		destIP := net.IP(buf[16:20])
		found := false
		for _, ip := range vpnNet.IPs {
			fmt.Println(ip)
			if ip == destIP.String() {
				found = true
				break
			}
		}
		if found {
			log.Printf("Got %d bytes | IP from: %s | IP dest: %s\n", n, THIS_IP, destIP.String())
		} else {
			log.Printf("Got %d bytes | IP from: %s to /dev/null", n, destIP.String())
		}

	}
}

func ReadSignalizer(conn *net.UDPConn, vpnNet *VpnNet) error {
	buf := make([]byte, SIGNALIZER_MTU)
	for {
		n, _, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			return err
		}
		recieved := buf[:n]
		fmt.Println("> ", string(recieved))
		err = json.Unmarshal(recieved, &vpnNet)
		if err != nil {
			return err
		}
	}
}

func main() {
	var vpnNet VpnNet
	wg := sync.WaitGroup{}
	wg.Add(2)

	tun, err := CreateTun()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Interface %s created | MTU %s | IP %s", INTERFACE_NAME, strconv.Itoa(MTU), THIS_IP)
	go ReadTun(tun, &vpnNet)

	conn, err := ConnectToSingalizer()
	if err != nil {
		log.Fatal(err)
	}
	go ReadSignalizer(conn, &vpnNet)

	netRequest := NetRequest{
		TunIP: THIS_IP,
		Type:  "add",
	}
	netRequestBytes, err := json.Marshal(netRequest)
	if err != nil {
		log.Fatal(err)
	}
	conn.Write(netRequestBytes)

	wg.Wait()

	conn.Close()

}
