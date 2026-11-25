package analysis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("resource not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Inputs(
	ctx context.Context,
	id int64,
) (Inputs, error) {
	var in Inputs

	err := r.db.QueryRow(ctx, `
		SELECT
			a.id,
			a.raw_snapshot,
			a.status,
			b.id,
			b.raw_snapshot,
			b.status,
			p.id,
			p.engine_policy
		FROM audits a
		JOIN hosts h
			ON h.id = a.host_id
		LEFT JOIN audits b
			ON b.id = h.baseline_audit_id
			AND b.host_id = h.id
		LEFT JOIN security_policies p
			ON p.id = h.policy_id
		WHERE a.id = $1
	`, id).Scan(
		&in.CurrentID,
		&in.Current,
		&in.CurrentStatus,
		&in.BaselineID,
		&in.Baseline,
		&in.BaselineStatus,
		&in.PolicyID,
		&in.Policy,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return in, ErrNotFound
	}
	return in, err
}

func (r *Repository) Save(
	ctx context.Context,
	run Run,
	result Result,
) (Run, error) {
	summary, err := json.Marshal(run.Summary)
	if err != nil {
		return run, err
	}

	raw, err := json.Marshal(result)
	if err != nil {
		return run, err
	}

	err = r.db.QueryRow(ctx, `
		INSERT INTO analysis_runs (
			current_audit_id,
			baseline_audit_id,
			policy_id,
			policy_snapshot,
			engine_version,
			summary,
			result
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`,
		run.CurrentID,
		run.BaselineID,
		run.PolicyID,
		run.Policy,
		run.EngineVersion,
		summary,
		raw,
	).Scan(&run.ID, &run.CreatedAt)

	return run, err
}

func (r *Repository) Get(
	ctx context.Context,
	id int64,
) (Run, error) {
	var run Run
	var summary []byte

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			current_audit_id,
			baseline_audit_id,
			policy_id,
			policy_snapshot,
			engine_version,
			summary,
			created_at
		FROM analysis_runs
		WHERE id = $1
	`, id).Scan(
		&run.ID,
		&run.CurrentID,
		&run.BaselineID,
		&run.PolicyID,
		&run.Policy,
		&run.EngineVersion,
		&summary,
		&run.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return run, ErrNotFound
	}
	if err != nil {
		return run, err
	}

	if err := json.Unmarshal(summary, &run.Summary); err != nil {
		return run, err
	}

	if run.Summary.PolicyEvaluation.Status == "" {
		run.Summary.PolicyEvaluation = PolicyEvaluation{
			Status: "not_recorded",
			Reason: "This analysis predates explicit policy evaluation status",
		}
	}

	return run, nil
}

func (r *Repository) Coverage(
	ctx context.Context,
	id int64,
) (json.RawMessage, error) {
	var raw json.RawMessage

	err := r.db.QueryRow(ctx, `
		SELECT result -> 'diff' -> 'coverage'
		FROM analysis_runs
		WHERE id = $1
	`, id).Scan(&raw)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return raw, err
}

func (r *Repository) Details(
	ctx context.Context,
	id int64,
	checks bool,
	filter Filter,
) (Page, error) {
	page := Page{
		Items:  []json.RawMessage{},
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	if _, err := r.Get(ctx, id); err != nil {
		return page, err
	}

	key := "changes"
	if checks {
		key = "checks"
	}

	rows, err := r.db.Query(ctx, `
		SELECT item
		FROM analysis_runs a
		CROSS JOIN LATERAL jsonb_array_elements(
			CASE
				WHEN $2 = 'checks' THEN a.result -> 'checks'
				ELSE a.result -> 'diff' -> 'changes'
			END
		) WITH ORDINALITY AS entries(item, position)
		WHERE a.id = $1
			AND ($3 = '' OR item ->> 'section' = $3)
			AND ($4 = '' OR item ->> 'kind' = $4)
			AND ($5 = '' OR item ->> 'status' = $5)
			AND ($6 = '' OR item ->> 'phase' = $6)
		ORDER BY position
		LIMIT $7 OFFSET $8
	`,
		id,
		key,
		filter.Section,
		filter.Kind,
		filter.Status,
		filter.Phase,
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return page, err
	}
	defer rows.Close()

	for rows.Next() {
		var item json.RawMessage
		if err := rows.Scan(&item); err != nil {
			return page, err
		}
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func (r *Repository) GetPolicy(
	ctx context.Context,
	id int64,
) (json.RawMessage, error) {
	var raw json.RawMessage

	err := r.db.QueryRow(ctx, `
		SELECT engine_policy
		FROM security_policies
		WHERE id = $1
	`, id).Scan(&raw)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err == nil && len(raw) == 0 {
		return nil, conflict("engine policy is not configured")
	}
	return raw, err
}

func (r *Repository) PutPolicy(
	ctx context.Context,
	id int64,
	raw json.RawMessage,
) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE security_policies
		SET engine_policy = $2,
			updated_at = NOW()
		WHERE id = $1
	`, id, raw)

	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}
