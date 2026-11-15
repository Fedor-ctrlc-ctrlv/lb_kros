package services

import (
	"fmt"
	"path"
	"strings"

	"web-api/internal/models"
)

func validateSnapshotSecurity(s *models.Snapshot) error {
	sec := s.Security

	if s.Version == "2.0" {
		if sec != nil {
			return fmt.Errorf("security requires snapshot version 2.1")
		}
		return nil
	}

	if sec == nil {
		return fmt.Errorf("security is required for snapshot version 2.1")
	}

	if len(sec.Scope.FilesystemRoots) == 0 {
		return fmt.Errorf("security.scope.filesystem_roots is required")
	}

	for _, root := range sec.Scope.FilesystemRoots {
		if !path.IsAbs(root) {
			return fmt.Errorf("filesystem root must be absolute: %q", root)
		}
	}

	collectors := []struct {
		name    string
		present bool
	}{
		{"listening_ports", sec.ListeningPorts != nil},
		{"ssh_keys", sec.SSHKeys != nil},
		{"firewall", sec.Firewall != nil},
		{"resolver", sec.Resolver != nil},
	}

	for _, collector := range collectors {
		switch sec.Status[collector.name] {
		case "ok", "partial":
			if !collector.present {
				return fmt.Errorf(
					"security.%s is required for its status",
					collector.name,
				)
			}
		case "unavailable":
			if collector.present {
				return fmt.Errorf(
					"security.%s must be null when unavailable",
					collector.name,
				)
			}
		default:
			return fmt.Errorf(
				"invalid or missing security.status.%s",
				collector.name,
			)
		}
	}

	for _, port := range sec.ListeningPorts {
		if (port.Protocol != "tcp" && port.Protocol != "udp") ||
			port.Port == 0 ||
			strings.TrimSpace(port.Address) == "" {
			return fmt.Errorf("invalid entry in security.listening_ports")
		}
	}

	for _, key := range sec.SSHKeys {
		if strings.TrimSpace(key.Username) == "" ||
			!path.IsAbs(key.Source) ||
			key.Type == "" ||
			!strings.HasPrefix(key.Fingerprint, "SHA256:") ||
			len(key.Fingerprint) == len("SHA256:") {
			return fmt.Errorf("invalid entry in security.ssh_keys")
		}
	}

	if firewall := sec.Firewall; firewall != nil &&
		(firewall.Backend != "nftables" || firewall.Ruleset == nil) {
		return fmt.Errorf(
			"security.firewall requires backend nftables and a ruleset array",
		)
	}

	return nil
}
