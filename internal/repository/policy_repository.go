package repository

import (
	"context"
	"errors"
	"fmt"

	"web-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPolicyNotFound = errors.New("policy not found")
	ErrRuleNotFound   = errors.New("rule not found")
)

type PolicyRepository struct {
	db *pgxpool.Pool
}

func NewPolicyRepository(db *pgxpool.Pool) *PolicyRepository {
	return &PolicyRepository{db: db}
}

func scanPolicy(row pgx.Row) (*models.SecurityPolicy, error) {
	var policy models.SecurityPolicy

	err := row.Scan(
		&policy.ID,
		&policy.Name,
		&policy.Description,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

func scanRule(row pgx.Row) (*models.PolicyRule, error) {
	var rule models.PolicyRule

	err := row.Scan(
		&rule.ID,
		&rule.PolicyID,
		&rule.Type,
		&rule.Enabled,
		&rule.Severity,
		&rule.Parameters,
	)
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func (r *PolicyRepository) Create(
	ctx context.Context,
	policy *models.SecurityPolicy,
) (*models.SecurityPolicy, error) {
	query := `
		INSERT INTO security_policies (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`

	result, err := scanPolicy(
		r.db.QueryRow(ctx, query, policy.Name, policy.Description),
	)
	if err != nil {
		return nil, fmt.Errorf("create policy: %w", err)
	}

	result.Rules = []models.PolicyRule{}
	return result, nil
}

func (r *PolicyRepository) GetByID(
	ctx context.Context,
	id int64,
) (*models.SecurityPolicy, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM security_policies
		WHERE id = $1
	`

	policy, err := scanPolicy(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPolicyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get policy: %w", err)
	}

	rules, err := r.ListRules(ctx, id)
	if err != nil {
		return nil, err
	}

	policy.Rules = rules
	return policy, nil
}

func (r *PolicyRepository) List(
	ctx context.Context,
) ([]models.SecurityPolicy, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM security_policies
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	defer rows.Close()

	policies := make([]models.SecurityPolicy, 0)

	for rows.Next() {
		policy, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		policy.Rules = []models.PolicyRule{}
		policies = append(policies, *policy)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return policies, nil
}

func (r *PolicyRepository) Update(
	ctx context.Context,
	policy *models.SecurityPolicy,
) (*models.SecurityPolicy, error) {
	query := `
		UPDATE security_policies
		SET name = $2, description = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at
	`

	result, err := scanPolicy(r.db.QueryRow(
		ctx, query, policy.ID, policy.Name, policy.Description,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPolicyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update policy: %w", err)
	}

	result.Rules, err = r.ListRules(ctx, result.ID)
	return result, err
}

func (r *PolicyRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM security_policies WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPolicyNotFound
	}
	return nil
}

func (r *PolicyRepository) ListRules(
	ctx context.Context,
	policyID int64,
) ([]models.PolicyRule, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, policy_id, rule_type, enabled, severity, parameters
		FROM policy_rules
		WHERE policy_id = $1
		ORDER BY id
	`, policyID)
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	defer rows.Close()

	rules := make([]models.PolicyRule, 0)

	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *PolicyRepository) CreateRule(
	ctx context.Context,
	rule *models.PolicyRule,
) (*models.PolicyRule, error) {
	query := `
		INSERT INTO policy_rules (
			policy_id, rule_type, enabled, severity, parameters
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, policy_id, rule_type, enabled, severity, parameters
	`

	result, err := scanRule(r.db.QueryRow(
		ctx, query,
		rule.PolicyID,
		rule.Type,
		rule.Enabled,
		rule.Severity,
		rule.Parameters,
	))
	if err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}

	return result, nil
}

func (r *PolicyRepository) GetRule(
	ctx context.Context,
	policyID, ruleID int64,
) (*models.PolicyRule, error) {
	query := `
		SELECT id, policy_id, rule_type, enabled, severity, parameters
		FROM policy_rules
		WHERE policy_id = $1 AND id = $2
	`

	rule, err := scanRule(
		r.db.QueryRow(ctx, query, policyID, ruleID),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get rule: %w", err)
	}

	return rule, nil
}

func (r *PolicyRepository) UpdateRule(
	ctx context.Context,
	rule *models.PolicyRule,
) (*models.PolicyRule, error) {
	query := `
		UPDATE policy_rules
		SET rule_type = $3,
		    enabled = $4,
		    severity = $5,
		    parameters = $6
		WHERE policy_id = $1 AND id = $2
		RETURNING id, policy_id, rule_type, enabled, severity, parameters
	`

	result, err := scanRule(r.db.QueryRow(
		ctx, query,
		rule.PolicyID,
		rule.ID,
		rule.Type,
		rule.Enabled,
		rule.Severity,
		rule.Parameters,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update rule: %w", err)
	}

	return result, nil
}

func (r *PolicyRepository) DeleteRule(
	ctx context.Context,
	policyID, ruleID int64,
) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM policy_rules
		WHERE policy_id = $1 AND id = $2
	`, policyID, ruleID)
	if err != nil {
		return fmt.Errorf("delete rule: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRuleNotFound
	}
	return nil
}
