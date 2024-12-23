package common

type ServiceType int

const (
	Invalid ServiceType = iota
	WebService
	LoadBalancer
	ServiceDiscovery
)

const Port = "8080"

type Heartbeat struct {
	Address           string
	LastHeartbeatTime int64
	Load              float64
	Success           bool
	ServiceType       ServiceType
}

type Host struct {
	Address string
	Load    float64
}

type HostList struct {
	Hosts []Host
}
