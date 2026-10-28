package repository

import (
	"context"
	"errors"
	"fmt"
	"web-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BaselineRepository struct {
	db *pgxpool.Pool
}

func NewBaselineRepository(db *pgxpool.Pool) *BaselineRepository {
	return &BaselineRepository{db: db}
}

func (r *BaselineRepository) Set(
	ctx context.Context,
	hostID, auditID int64,
) (*AuditRecord, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int64
	err = tx.QueryRow(ctx,
		`SELECT id FROM hosts WHERE id = $1 FOR UPDATE`,
		hostID,
	).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrHostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find host: %w", err)
	}

	var auditHostID int64
	var status string

	err = tx.QueryRow(ctx, `
		SELECT host_id, status
		FROM audits
		WHERE id = $1
	`, auditID).Scan(&auditHostID, &status)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrAuditNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find audit: %w", err)
	}

	if auditHostID != hostID {
		return nil, models.ErrAuditHostMismatch
	}
	if status != "COMPLETED" {
		return nil, models.ErrAuditNotCompleted
	}

	_, err = tx.Exec(ctx, `
		UPDATE hosts
		SET baseline_audit_id = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, hostID, auditID)
	if err != nil {
		return nil, fmt.Errorf("set baseline: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit baseline: %w", err)
	}

	return r.Get(ctx, hostID)
}

func (r *BaselineRepository) Get(
	ctx context.Context,
	hostID int64,
) (*AuditRecord, error) {
	var baselineID *int64

	err := r.db.QueryRow(ctx, `
		SELECT baseline_audit_id
		FROM hosts
		WHERE id = $1
	`, hostID).Scan(&baselineID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrHostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get host baseline: %w", err)
	}

	if baselineID == nil {
		return nil, models.ErrBaselineNotSet
	}

	audit, err := scanAudit(r.db.QueryRow(ctx,
		`SELECT `+auditColumns+`
		 FROM audits
		 WHERE id = $1 AND host_id = $2`,
		*baselineID, hostID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrAuditNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get baseline audit: %w", err)
	}

	return audit, nil
}

func (r *BaselineRepository) Clear(
	ctx context.Context,
	hostID int64,
) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE hosts
		SET baseline_audit_id = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`, hostID)

	if err != nil {
		return fmt.Errorf("clear baseline: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrHostNotFound
	}

	return nil
}
