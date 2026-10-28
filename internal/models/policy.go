package models

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type SecurityPolicy struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Rules []PolicyRule `json:"rules"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PolicyRule struct {
	ID       int64 `json:"id"`
	PolicyID int64 `json:"policy_id"`

	Type     RuleType `json:"type"`
	Severity Severity `json:"severity"`
	Enabled  bool     `json:"enabled"`

	Parameters json.RawMessage `json:"parameters"`
}

func (p SecurityPolicy) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("policy name cannot be empty")
	}

	seen := make(map[RuleType]bool)

	for _, rule := range p.Rules {
		if err := rule.Validate(); err != nil {
			return err
		}

		if seen[rule.Type] {
			return errors.New("duplicate rule type")
		}

		seen[rule.Type] = true
	}

	return nil
}

func (p *SecurityPolicy) AddRule(rule PolicyRule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	for _, existing := range p.Rules {
		if existing.Type == rule.Type {
			return errors.New("rule type already exists")
		}
	}

	rule.PolicyID = p.ID

	p.Rules = append(p.Rules, rule)
	p.UpdatedAt = time.Now().UTC()

	return nil
}

func (p *SecurityPolicy) RemoveRule(ruleType RuleType) error {
	for i, rule := range p.Rules {
		if rule.Type == ruleType {
			p.Rules = append(
				p.Rules[:i],
				p.Rules[i+1:]...,
			)

			p.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return errors.New("rule not found")
}

func (p *SecurityPolicy) FindRule(
	ruleType RuleType,
) (*PolicyRule, error) {
	for i := range p.Rules {
		if p.Rules[i].Type == ruleType {
			return &p.Rules[i], nil
		}
	}

	return nil, errors.New("rule not found")
}

func (p *SecurityPolicy) EnableRule(
	ruleType RuleType,
) error {
	rule, err := p.FindRule(ruleType)
	if err != nil {
		return err
	}

	rule.Enable()
	p.UpdatedAt = time.Now().UTC()

	return nil
}

func (p *SecurityPolicy) DisableRule(
	ruleType RuleType,
) error {
	rule, err := p.FindRule(ruleType)
	if err != nil {
		return err
	}

	rule.Disable()
	p.UpdatedAt = time.Now().UTC()

	return nil
}

func (p SecurityPolicy) ActiveRules() []PolicyRule {
	result := make([]PolicyRule, 0)

	for _, rule := range p.Rules {
		if rule.Enabled {
			result = append(result, rule)
		}
	}

	return result
}

func (r PolicyRule) Validate() error {
	switch string(r.Type) {
	case "USER_ADDED", "USER_REMOVED",
		"UNAUTHORIZED_UID_ZERO", "USER_ADDED_TO_SUDO",
		"SUID_ADDED", "SGID_ADDED",
		"CRITICAL_FILE_CHANGED", "WORLD_WRITABLE_FILE",
		"LISTENING_PORT_ADDED", "UNAUTHORIZED_PORT",
		"SERVICE_ADDED", "SERVICE_REMOVED",
		"CRON_JOB_ADDED", "SSH_KEY_ADDED", "SSH_KEY_REMOVED":
	default:
		return errors.New("invalid rule type")
	}

	switch r.Severity {
	case SeverityLow,
		SeverityMedium,
		SeverityHigh,
		SeverityCritical:
	default:
		return errors.New("invalid rule severity")
	}

	var parameters map[string]json.RawMessage

	if err := json.Unmarshal(r.Parameters, &parameters); err != nil {
		return errors.New("invalid rule parameters")
	}

	if parameters == nil {
		return errors.New("rule parameters must be a JSON object")
	}

	return nil
}

func (r *PolicyRule) Enable() {
	r.Enabled = true
}

func (r *PolicyRule) Disable() {
	r.Enabled = false
}

func (r *PolicyRule) SetSeverity(
	severity Severity,
) error {
	switch severity {
	case SeverityInfo,
		SeverityLow,
		SeverityMedium,
		SeverityHigh,
		SeverityCritical:
		r.Severity = severity
		return nil
	default:
		return errors.New("invalid severity")
	}
}

func (r *PolicyRule) SetParameters(
	parameters json.RawMessage,
) error {
	if len(parameters) > 0 &&
		!json.Valid(parameters) {
		return errors.New("invalid JSON parameters")
	}

	r.Parameters = append(
		json.RawMessage(nil),
		parameters...,
	)

	return nil
}
