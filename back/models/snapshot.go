package models

import (
	"encoding/json"
	"time"
)

type Snapshot struct {
	Version    string            `json:"version"`
	CreatedAt  time.Time         `json:"created_at"`
	System     SnapshotSystem    `json:"system"`
	Files      []SnapshotFile    `json:"files"`
	Users      []SnapshotUser    `json:"users"`
	Groups     []SnapshotGroup   `json:"groups"`
	Services   []SnapshotService `json:"services"`
	CronJobs   []SnapshotCron    `json:"cron_jobs"`
	ScanErrors json.RawMessage   `json:"scan_errors"`
}

type SnapshotSystem struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Kernel       string `json:"kernel"`
}

type SnapshotFile struct {
	Path        string     `json:"path"`
	Type        string     `json:"type"`
	LinkTarget  string     `json:"link_target"`
	SHA256      string     `json:"sha256"`
	HashStatus  string     `json:"hash_status"`
	Permissions string     `json:"permissions"`
	Mode        int64      `json:"mode"`
	UID         int64      `json:"uid"`
	GID         int64      `json:"gid"`
	Size        int64      `json:"size"`
	ModifiedAt  *time.Time `json:"modified_at"`
	SUID        bool       `json:"suid"`
	SGID        bool       `json:"sgid"`
	Sticky      bool       `json:"sticky"`
}

type SnapshotUser struct {
	Username string `json:"username"`
	UID      int64  `json:"uid"`
	GID      int64  `json:"gid"`
	Home     string `json:"home"`
	Shell    string `json:"shell"`
}

type SnapshotGroup struct {
	Name    string   `json:"name"`
	GID     int64    `json:"gid"`
	Members []string `json:"members"`
}

type SnapshotService struct {
	Name        string `json:"name"`
	LoadState   string `json:"load_state"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
}

type SnapshotCron struct {
	Source  string `json:"source"`
	Content string `json:"content"`
}
