package repository

import (
	"context"
	"errors"
	"fmt"

	"web-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PolicyRepository struct {
	db *pgxpool.Pool
}

func NewPolicyRepository(db *pgxpool.Pool) *PolicyRepository {
	return &PolicyRepository{db: db}
}

const policyColumns = `
	id,
	name,
	description,
	engine_policy,
	(engine_policy IS NOT NULL),
	COALESCE(
		jsonb_array_length(
			NULLIF(engine_policy -> 'state_rules', 'null'::jsonb)
		),
		0
	),
	COALESCE(
		jsonb_array_length(
			NULLIF(engine_policy -> 'change_rules', 'null'::jsonb)
		),
		0
	),
	created_at,
	updated_at
`

func scanPolicy(row pgx.Row) (*models.SecurityPolicy, error) {
	var policy models.SecurityPolicy

	err := row.Scan(
		&policy.ID,
		&policy.Name,
		&policy.Description,
		&policy.EnginePolicy,
		&policy.Configured,
		&policy.StateRuleCount,
		&policy.ChangeRuleCount,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

func (r *PolicyRepository) Create(
	ctx context.Context,
	policy *models.SecurityPolicy,
) (*models.SecurityPolicy, error) {
	result, err := scanPolicy(r.db.QueryRow(ctx, `
		INSERT INTO security_policies (
			name,
			description
		)
		VALUES ($1, $2)
		RETURNING `+policyColumns,
		policy.Name,
		policy.Description,
	))
	if err != nil {
		return nil, fmt.Errorf("create policy: %w", err)
	}

	return result, nil
}

func (r *PolicyRepository) GetByID(
	ctx context.Context,
	id int64,
) (*models.SecurityPolicy, error) {
	result, err := scanPolicy(r.db.QueryRow(ctx,
		`SELECT `+policyColumns+`
		 FROM security_policies
		 WHERE id = $1`,
		id,
	))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrPolicyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get policy: %w", err)
	}

	return result, nil
}

func (r *PolicyRepository) List(
	ctx context.Context,
) ([]models.SecurityPolicy, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+policyColumns+`
		 FROM security_policies
		 ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	defer rows.Close()

	result := []models.SecurityPolicy{}

	for rows.Next() {
		policy, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, *policy)
	}

	return result, rows.Err()
}

func (r *PolicyRepository) Update(
	ctx context.Context,
	policy *models.SecurityPolicy,
) (*models.SecurityPolicy, error) {
	result, err := scanPolicy(r.db.QueryRow(ctx, `
		UPDATE security_policies
		SET name = $2,
			description = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING `+policyColumns,
		policy.ID,
		policy.Name,
		policy.Description,
	))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrPolicyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update policy: %w", err)
	}

	return result, nil
}

func (r *PolicyRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM security_policies WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrPolicyNotFound
	}

	return nil
}
