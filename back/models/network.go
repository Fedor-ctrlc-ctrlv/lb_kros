package models

type NetworkEndpoint struct {
	ID      int64 `json:"id"`
	AuditID int64 `json:"audit_id"`

	Protocol string `json:"protocol"`

	LocalAddress string `json:"local_address"`
	LocalPort    uint16 `json:"local_port"`

	RemoteAddress string `json:"remote_address"`
	RemotePort    uint16 `json:"remote_port"`

	State string `json:"state"`

	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
}

func (n NetworkEndpoint) IsListening() bool {
	return n.State == "LISTEN"
}

func (n NetworkEndpoint) IsPublicBind() bool {
	return n.LocalAddress == "0.0.0.0" ||
		n.LocalAddress == "::"
}

func (n NetworkEndpoint) UsesPort(port uint16) bool {
	return n.LocalPort == port
}

func (n NetworkEndpoint) SameLocalEndpoint(other NetworkEndpoint) bool {
	return n.Protocol == other.Protocol &&
		n.LocalAddress == other.LocalAddress &&
		n.LocalPort == other.LocalPort
}
