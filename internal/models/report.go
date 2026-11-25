package models

import (
	"encoding/json"
	"time"
)

type HostPolicyReport struct {
	HostID     int64   `json:"host_id"`
	HostName   string  `json:"host_name"`
	Hostname   string  `json:"hostname"`
	PolicyID   *int64  `json:"policy_id"`
	PolicyName *string `json:"policy_name"`
}

type HostBaselineReport struct {
	HostID            int64      `json:"host_id"`
	HostName          string     `json:"host_name"`
	Hostname          string     `json:"hostname"`
	BaselineAuditID   *int64     `json:"baseline_audit_id"`
	BaselineStatus    *string    `json:"baseline_status"`
	BaselineCreatedAt *time.Time `json:"baseline_created_at"`
}

type AuditHostReport struct {
	AuditID           int64     `json:"audit_id"`
	HostID            int64     `json:"host_id"`
	HostName          string    `json:"host_name"`
	Hostname          string    `json:"hostname"`
	Status            string    `json:"status"`
	SnapshotVersion   string    `json:"snapshot_version"`
	SnapshotCreatedAt time.Time `json:"snapshot_created_at"`
}

type PolicyRuleReport struct {
	PolicyID   int64           `json:"policy_id"`
	PolicyName string          `json:"policy_name"`
	Phase      string          `json:"phase"`
	RuleID     string          `json:"rule_id"`
	Section    string          `json:"section"`
	Definition json.RawMessage `json:"definition"`
}
