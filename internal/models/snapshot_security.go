package models

import "encoding/json"

type SnapshotSecurity struct {
	Scope          SnapshotScope     `json:"scope"`
	Status         map[string]string `json:"status"`
	ListeningPorts []SnapshotPort    `json:"listening_ports"`
	SSHKeys        []SnapshotSSHKey  `json:"ssh_keys"`
	Firewall       *SnapshotFirewall `json:"firewall"`
	Resolver       *SnapshotResolver `json:"resolver"`
}

type SnapshotScope struct {
	FilesystemRoots []string `json:"filesystem_roots"`
}

type SnapshotPort struct {
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     uint16 `json:"port"`
}

type SnapshotSSHKey struct {
	Username    string `json:"username"`
	Source      string `json:"source"`
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
}

type SnapshotFirewall struct {
	Backend string            `json:"backend"`
	Ruleset []json.RawMessage `json:"ruleset"`
}

type SnapshotResolver struct {
	LinkTarget  string   `json:"link_target,omitempty"`
	Nameservers []string `json:"nameservers"`
	Search      []string `json:"search"`
	Options     []string `json:"options"`
}
