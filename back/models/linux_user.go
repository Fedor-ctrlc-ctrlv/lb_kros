package models

import "strings"

type LinuxUser struct {
	ID      int64 `json:"id"`
	AuditID int64 `json:"audit_id"`

	Username string `json:"username"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`

	Home  string `json:"home"`
	Shell string `json:"shell"`

	Groups []string `json:"groups"`
	IsSudo bool     `json:"is_sudo"`
}

func (u LinuxUser) IsRoot() bool {
	return u.UID == 0
}

func (u LinuxUser) HasLoginShell() bool {
	shell := strings.TrimSpace(u.Shell)

	return shell != "" &&
		shell != "/usr/sbin/nologin" &&
		shell != "/sbin/nologin" &&
		shell != "/bin/false"
}

func (u LinuxUser) BelongsToGroup(name string) bool {
	for _, group := range u.Groups {
		if group == name {
			return true
		}
	}

	return false
}

func (u LinuxUser) IsPrivileged() bool {
	return u.IsRoot() || u.IsSudo
}
