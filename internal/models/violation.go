package models

import (
	"errors"
	"strings"
	"time"
)

type Violation struct {
	ID int64 `json:"id"`

	AuditID int64 `json:"audit_id"`
	RuleID  int64 `json:"rule_id"`

	RuleType RuleType `json:"rule_type"`
	Severity Severity `json:"severity"`

	Title       string `json:"title"`
	Description string `json:"description"`

	EntityType string `json:"entity_type"`
	EntityKey  string `json:"entity_key"`

	CreatedAt time.Time `json:"created_at"`
}

func (v Violation) IsCritical() bool {
	return v.Severity == SeverityCritical
}

func (v Violation) IsHighPriority() bool {
	return v.Severity == SeverityCritical ||
		v.Severity == SeverityHigh
}

func (v Violation) Validate() error {
	if v.AuditID <= 0 {
		return errors.New("invalid audit id")
	}

	if v.RuleType == "" {
		return errors.New("rule type cannot be empty")
	}

	if strings.TrimSpace(v.Title) == "" {
		return errors.New("violation title cannot be empty")
	}

	return nil
}
