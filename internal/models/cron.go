package models

import "strings"

type CronJob struct {
	ID      int64 `json:"id"`
	AuditID int64 `json:"audit_id"`

	User string `json:"user"`

	Schedule string `json:"schedule"`
	Command  string `json:"command"`

	Source string `json:"source"`
}

func (c CronJob) IsRootJob() bool {
	return c.User == "root"
}

func (c CronJob) IsRebootJob() bool {
	return strings.HasPrefix(
		strings.TrimSpace(c.Schedule),
		"@reboot",
	)
}

func (c CronJob) SameJob(other CronJob) bool {
	return c.User == other.User &&
		c.Schedule == other.Schedule &&
		c.Command == other.Command &&
		c.Source == other.Source
}
