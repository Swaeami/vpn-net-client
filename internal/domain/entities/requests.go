package entities

type VpnNet struct {
	Name string
	IPs  []string
}

type NetRequest struct {
	TunIP string
	Type  string
}
