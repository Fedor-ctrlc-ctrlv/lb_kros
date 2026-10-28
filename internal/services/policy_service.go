package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"web-api/internal/models"
)

var ErrInvalidPolicy = errors.New("invalid policy data")
var ErrInvalidRule = errors.New("invalid policy rule")

type PolicyStore interface {
	Create(context.Context, *models.SecurityPolicy) (*models.SecurityPolicy, error)
	GetByID(context.Context, int64) (*models.SecurityPolicy, error)
	List(context.Context) ([]models.SecurityPolicy, error)
	Update(context.Context, *models.SecurityPolicy) (*models.SecurityPolicy, error)
	Delete(context.Context, int64) error
	CreateRule(context.Context, *models.PolicyRule) (*models.PolicyRule, error)
	GetRule(context.Context, int64, int64) (*models.PolicyRule, error)
	UpdateRule(context.Context, *models.PolicyRule) (*models.PolicyRule, error)
	DeleteRule(context.Context, int64, int64) error
}

type PolicyService struct {
	repo PolicyStore
}

func NewPolicyService(repo PolicyStore) *PolicyService {
	return &PolicyService{repo: repo}
}

type CreatePolicyInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatePolicyInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateRuleInput struct {
	Type       models.RuleType `json:"type"`
	Severity   models.Severity `json:"severity"`
	Enabled    *bool           `json:"enabled"`
	Parameters json.RawMessage `json:"parameters"`
}

type UpdateRuleInput struct {
	Type       *models.RuleType `json:"type"`
	Severity   *models.Severity `json:"severity"`
	Enabled    *bool            `json:"enabled"`
	Parameters json.RawMessage  `json:"parameters"`
}

func validSeverity(s models.Severity) bool {
	switch string(s) {
	case "LOW", "MEDIUM", "HIGH", "CRITICAL":
		return true
	default:
		return false
	}
}

func validRuleType(t models.RuleType) bool {
	switch string(t) {
	case "USER_ADDED", "USER_REMOVED",
		"UNAUTHORIZED_UID_ZERO", "USER_ADDED_TO_SUDO",
		"SUID_ADDED", "SGID_ADDED",
		"CRITICAL_FILE_CHANGED", "WORLD_WRITABLE_FILE",
		"LISTENING_PORT_ADDED", "UNAUTHORIZED_PORT",
		"SERVICE_ADDED", "SERVICE_REMOVED",
		"CRON_JOB_ADDED", "SSH_KEY_ADDED", "SSH_KEY_REMOVED":
		return true
	default:
		return false
	}
}

func validParameters(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return false
	}

	return obj != nil
}

func (s *PolicyService) Create(
	ctx context.Context,
	input CreatePolicyInput,
) (*models.SecurityPolicy, error) {
	policy := &models.SecurityPolicy{
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
	}
	if policy.Name == "" {
		return nil, ErrInvalidPolicy
	}
	return s.repo.Create(ctx, policy)
}

func (s *PolicyService) GetByID(
	ctx context.Context,
	id int64,
) (*models.SecurityPolicy, error) {
	if id <= 0 {
		return nil, ErrInvalidPolicy
	}
	return s.repo.GetByID(ctx, id)
}

func (s *PolicyService) List(
	ctx context.Context,
) ([]models.SecurityPolicy, error) {
	return s.repo.List(ctx)
}

func (s *PolicyService) Update(
	ctx context.Context,
	id int64,
	input UpdatePolicyInput,
) (*models.SecurityPolicy, error) {
	if id <= 0 {
		return nil, ErrInvalidPolicy
	}

	policy, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		policy.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		policy.Description = strings.TrimSpace(*input.Description)
	}
	if policy.Name == "" {
		return nil, ErrInvalidPolicy
	}

	return s.repo.Update(ctx, policy)
}

func (s *PolicyService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidPolicy
	}
	return s.repo.Delete(ctx, id)
}

func (s *PolicyService) CreateRule(
	ctx context.Context,
	policyID int64,
	input CreateRuleInput,
) (*models.PolicyRule, error) {
	if policyID <= 0 ||
		!validRuleType(input.Type) ||
		!validSeverity(input.Severity) {
		return nil, ErrInvalidRule
	}

	parameters := input.Parameters
	if len(parameters) == 0 {
		parameters = json.RawMessage(`{}`)
	}
	if !validParameters(parameters) {
		return nil, ErrInvalidRule
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	if _, err := s.repo.GetByID(ctx, policyID); err != nil {
		return nil, err
	}

	rule := &models.PolicyRule{
		PolicyID:   policyID,
		Type:       input.Type,
		Severity:   input.Severity,
		Enabled:    enabled,
		Parameters: parameters,
	}

	return s.repo.CreateRule(ctx, rule)
}

func (s *PolicyService) UpdateRule(
	ctx context.Context,
	policyID, ruleID int64,
	input UpdateRuleInput,
) (*models.PolicyRule, error) {
	if policyID <= 0 || ruleID <= 0 {
		return nil, ErrInvalidRule
	}

	rule, err := s.repo.GetRule(ctx, policyID, ruleID)
	if err != nil {
		return nil, err
	}

	if input.Type != nil {
		rule.Type = *input.Type
	}
	if input.Severity != nil {
		rule.Severity = *input.Severity
	}
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}
	if input.Parameters != nil {
		rule.Parameters = input.Parameters
	}

	if !validRuleType(rule.Type) ||
		!validSeverity(rule.Severity) ||
		!validParameters(rule.Parameters) {
		return nil, ErrInvalidRule
	}

	return s.repo.UpdateRule(ctx, rule)
}

func (s *PolicyService) DeleteRule(
	ctx context.Context,
	policyID, ruleID int64,
) error {
	if policyID <= 0 || ruleID <= 0 {
		return ErrInvalidRule
	}
	return s.repo.DeleteRule(ctx, policyID, ruleID)
}
