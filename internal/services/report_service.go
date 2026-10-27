package services

import (
	"context"
	"errors"
	"strings"

	"web-api/internal/models"
)

var ErrInvalidSeverity = errors.New("invalid severity")

type ReportStore interface {
	HostsWithPolicies(context.Context) ([]models.HostPolicyReport, error)
	HostsWithBaselines(context.Context) ([]models.HostBaselineReport, error)
	AuditsWithHosts(context.Context) ([]models.AuditHostReport, error)
	PolicyRules(context.Context) ([]models.PolicyRuleReport, error)
}

type ReportService struct {
	repo ReportStore
}

func NewReportService(repo ReportStore) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) HostsWithPolicies(
	ctx context.Context,
) ([]models.HostPolicyReport, error) {
	return s.repo.HostsWithPolicies(ctx)
}

func (s *ReportService) HostsWithBaselines(
	ctx context.Context,
) ([]models.HostBaselineReport, error) {
	return s.repo.HostsWithBaselines(ctx)
}

func (s *ReportService) AuditsWithHosts(
	ctx context.Context,
) ([]models.AuditHostReport, error) {
	return s.repo.AuditsWithHosts(ctx)
}

type PolicyRuleView struct {
	PolicyName string `json:"policy_name"`
	RuleType   string `json:"rule_type"`
	Severity   string `json:"severity"`
	Enabled    bool   `json:"enabled"`
}

func (s *ReportService) PolicyRules(
	ctx context.Context,
	severity string,
) ([]PolicyRuleView, error) {
	severity = strings.ToUpper(strings.TrimSpace(severity))

	if severity != "" &&
		severity != "LOW" &&
		severity != "MEDIUM" &&
		severity != "HIGH" &&
		severity != "CRITICAL" {
		return nil, ErrInvalidSeverity
	}

	rules, err := s.repo.PolicyRules(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]PolicyRuleView, 0)

	for _, rule := range rules {
		// Аналог LINQ
		if severity != "" && string(rule.Severity) != severity {
			continue
		}

		// Аналог LINQ для лабы...
		result = append(result, PolicyRuleView{
			PolicyName: rule.PolicyName,
			RuleType:   string(rule.RuleType),
			Severity:   string(rule.Severity),
			Enabled:    rule.Enabled,
		})
	}

	return result, nil
}
