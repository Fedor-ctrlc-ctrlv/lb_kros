package models

import (
	"errors"
	"strings"
	"time"
)

type Host struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Hostname    string `json:"hostname"`
	Description string `json:"description"`

	OS           string `json:"os"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`

	PolicyID        *int64 `json:"policy_id"`
	BaselineAuditID *int64 `json:"baseline_audit_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *Host) AssignPolicy(policyID int64) error {
	if policyID <= 0 {
		return errors.New("invalid policy id")
	}

	h.PolicyID = &policyID
	h.UpdatedAt = time.Now().UTC()
	return nil
}

func (h *Host) RemovePolicy() {
	h.PolicyID = nil
	h.UpdatedAt = time.Now().UTC()
}

func (h *Host) SetBaseline(auditID int64) error {
	if auditID <= 0 {
		return errors.New("invalid audit id")
	}

	h.BaselineAuditID = &auditID
	h.UpdatedAt = time.Now().UTC()
	return nil
}

func (h *Host) ClearBaseline() {
	h.BaselineAuditID = nil
	h.UpdatedAt = time.Now().UTC()
}

func (h Host) HasBaseline() bool {
	return h.BaselineAuditID != nil
}

func (h Host) Validate() error {
	if strings.TrimSpace(h.Name) == "" {
		return errors.New("host name cannot be empty")
	}

	if strings.TrimSpace(h.Hostname) == "" {
		return errors.New("hostname cannot be empty")
	}

	return nil
}
