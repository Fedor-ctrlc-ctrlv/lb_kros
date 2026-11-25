package services

import (
	"context"
	"errors"
	"strings"

	"web-api/internal/models"
)

var ErrInvalidReportFilter = errors.New("invalid report filter")

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

func (s *ReportService) PolicyRules(
	ctx context.Context,
	phase string,
) ([]models.PolicyRuleReport, error) {
	phase = strings.ToLower(strings.TrimSpace(phase))

	if phase != "" && phase != "state" && phase != "change" {
		return nil, ErrInvalidReportFilter
	}

	rules, err := s.repo.PolicyRules(ctx)
	if err != nil {
		return nil, err
	}

	result := []models.PolicyRuleReport{}

	for _, rule := range rules {
		if phase != "" && rule.Phase != phase {
			continue
		}

		result = append(result, rule)
	}

	return result, nil
}
