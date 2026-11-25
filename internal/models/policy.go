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

	EnginePolicy json.RawMessage `json:"engine_policy"`

	Configured      bool `json:"configured"`
	StateRuleCount  int  `json:"state_rule_count"`
	ChangeRuleCount int  `json:"change_rule_count"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p SecurityPolicy) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("policy name cannot be empty")
	}
	return nil
}
