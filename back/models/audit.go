package models

import (
	"errors"
	"time"
)

type Audit struct {
	ID     int64 `json:"id"`
	HostID int64 `json:"host_id"`

	Status AuditStatus `json:"status"`

	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`

	CollectorVersion string `json:"collector_version"`

	LinuxUsers       []LinuxUser       `json:"linux_users"`
	LinuxGroups      []LinuxGroup      `json:"linux_groups"`
	FsNodes          []FsNode          `json:"fs_nodes"`
	Services         []Service         `json:"services"`
	NetworkEndpoints []NetworkEndpoint `json:"network_endpoints"`
	CronJobs         []CronJob         `json:"cron_jobs"`
	SSHKeys          []SSHKey          `json:"ssh_keys"`

	CreatedAt time.Time `json:"created_at"`
}

func NewAudit(hostID int64) (*Audit, error) {
	if hostID <= 0 {
		return nil, errors.New("invalid host id")
	}

	now := time.Now().UTC()

	return &Audit{
		HostID:    hostID,
		Status:    AuditPending,
		StartedAt: now,
		CreatedAt: now,
	}, nil
}

func (a *Audit) Complete() error {
	if a.Status != AuditPending {
		return errors.New("only pending audit can be completed")
	}

	now := time.Now().UTC()

	a.Status = AuditCompleted
	a.FinishedAt = &now

	return nil
}

func (a *Audit) Fail() error {
	if a.Status != AuditPending {
		return errors.New("only pending audit can be failed")
	}

	now := time.Now().UTC()

	a.Status = AuditFailed
	a.FinishedAt = &now

	return nil
}

func (a Audit) IsCompleted() bool {
	return a.Status == AuditCompleted
}

func (a Audit) TotalObjects() int {
	return len(a.LinuxUsers) +
		len(a.LinuxGroups) +
		len(a.FsNodes) +
		len(a.Services) +
		len(a.NetworkEndpoints) +
		len(a.CronJobs) +
		len(a.SSHKeys)
}

func (a Audit) Duration() time.Duration {
	if a.FinishedAt == nil {
		return 0
	}

	return a.FinishedAt.Sub(a.StartedAt)
}
