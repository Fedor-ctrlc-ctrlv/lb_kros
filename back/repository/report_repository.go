package repository

import (
	"context"
	"fmt"

	"web-api/back/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) HostsWithPolicies(
	ctx context.Context,
) ([]models.HostPolicyReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			h.id,
			h.name,
			h.hostname,
			p.id,
			p.name
		FROM hosts h
		LEFT JOIN security_policies p
			ON h.policy_id = p.id
		ORDER BY h.id
	`)
	if err != nil {
		return nil, fmt.Errorf("query hosts with policies: %w", err)
	}
	defer rows.Close()

	result := make([]models.HostPolicyReport, 0)

	for rows.Next() {
		var item models.HostPolicyReport

		if err := rows.Scan(
			&item.HostID,
			&item.HostName,
			&item.Hostname,
			&item.PolicyID,
			&item.PolicyName,
		); err != nil {
			return nil, fmt.Errorf("scan host policy: %w", err)
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *ReportRepository) HostsWithBaselines(
	ctx context.Context,
) ([]models.HostBaselineReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			h.id,
			h.name,
			h.hostname,
			a.id,
			a.status,
			a.snapshot_created_at
		FROM hosts h
		LEFT JOIN audits a
			ON a.id = h.baseline_audit_id
			AND a.host_id = h.id
		ORDER BY h.id
	`)
	if err != nil {
		return nil, fmt.Errorf("query hosts with baselines: %w", err)
	}
	defer rows.Close()

	result := make([]models.HostBaselineReport, 0)

	for rows.Next() {
		var item models.HostBaselineReport

		if err := rows.Scan(
			&item.HostID,
			&item.HostName,
			&item.Hostname,
			&item.BaselineAuditID,
			&item.BaselineStatus,
			&item.BaselineCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan host baseline: %w", err)
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *ReportRepository) AuditsWithHosts(
	ctx context.Context,
) ([]models.AuditHostReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			a.id,
			h.id,
			h.name,
			h.hostname,
			a.status,
			COALESCE(a.snapshot_version, ''),
			a.snapshot_created_at
		FROM audits a
		INNER JOIN hosts h
			ON a.host_id = h.id
		ORDER BY a.id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query audits with hosts: %w", err)
	}
	defer rows.Close()

	result := make([]models.AuditHostReport, 0)

	for rows.Next() {
		var item models.AuditHostReport

		if err := rows.Scan(
			&item.AuditID,
			&item.HostID,
			&item.HostName,
			&item.Hostname,
			&item.Status,
			&item.SnapshotVersion,
			&item.SnapshotCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit host: %w", err)
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *ReportRepository) PolicyRules(
	ctx context.Context,
) ([]models.PolicyRuleReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			p.id,
			p.name,
			pr.id,
			pr.rule_type,
			pr.severity,
			pr.enabled
		FROM security_policies p
		INNER JOIN policy_rules pr
			ON pr.policy_id = p.id
		ORDER BY p.id, pr.id
	`)
	if err != nil {
		return nil, fmt.Errorf("query policy rules: %w", err)
	}
	defer rows.Close()

	result := make([]models.PolicyRuleReport, 0)

	for rows.Next() {
		var item models.PolicyRuleReport

		if err := rows.Scan(
			&item.PolicyID,
			&item.PolicyName,
			&item.RuleID,
			&item.RuleType,
			&item.Severity,
			&item.Enabled,
		); err != nil {
			return nil, fmt.Errorf("scan policy rule: %w", err)
		}

		result = append(result, item)
	}

	return result, rows.Err()
}
