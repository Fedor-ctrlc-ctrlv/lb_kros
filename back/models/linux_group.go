package models

type LinuxGroup struct {
	ID      int64 `json:"id"`
	AuditID int64 `json:"audit_id"`

	Name    string   `json:"name"`
	GID     int      `json:"gid"`
	Members []string `json:"members"`
}

func (g LinuxGroup) HasMember(username string) bool {
	for _, member := range g.Members {
		if member == username {
			return true
		}
	}

	return false
}

func (g LinuxGroup) MemberCount() int {
	return len(g.Members)
}

func (g LinuxGroup) IsPrivilegedGroup() bool {
	return g.Name == "root" ||
		g.Name == "sudo" ||
		g.Name == "wheel"
}
