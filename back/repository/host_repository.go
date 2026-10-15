package repository

import (
	"context"
	"errors"
	"fmt"

	"web-api/back/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrHostNotFound = errors.New("host not found")

type HostRepository struct {
	db *pgxpool.Pool
}

func NewHostRepository(db *pgxpool.Pool) *HostRepository {
	return &HostRepository{db: db}
}

const hostColumns = `
	id, name, hostname, description, os,
	architecture, kernel, policy_id,
	baseline_audit_id, created_at, updated_at
`

func scanHost(row pgx.Row) (*models.Host, error) {
	var host models.Host

	err := row.Scan(
		&host.ID,
		&host.Name,
		&host.Hostname,
		&host.Description,
		&host.OS,
		&host.Architecture,
		&host.Kernel,
		&host.PolicyID,
		&host.BaselineAuditID,
		&host.CreatedAt,
		&host.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &host, nil
}

func (r *HostRepository) Create(
	ctx context.Context,
	host *models.Host,
) (*models.Host, error) {
	query := `
		INSERT INTO hosts (
			name, hostname, description, os,
			architecture, kernel
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + hostColumns

	row := r.db.QueryRow(
		ctx, query,
		host.Name,
		host.Hostname,
		host.Description,
		host.OS,
		host.Architecture,
		host.Kernel,
	)

	result, err := scanHost(row)
	if err != nil {
		return nil, fmt.Errorf("create host: %w", err)
	}

	return result, nil
}

func (r *HostRepository) GetByID(
	ctx context.Context,
	id int64,
) (*models.Host, error) {
	query := `
		SELECT ` + hostColumns + `
		FROM hosts
		WHERE id = $1
	`

	host, err := scanHost(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrHostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get host: %w", err)
	}

	return host, nil
}

func (r *HostRepository) List(
	ctx context.Context,
) ([]models.Host, error) {
	query := `
		SELECT ` + hostColumns + `
		FROM hosts
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list hosts: %w", err)
	}
	defer rows.Close()

	hosts := make([]models.Host, 0)

	for rows.Next() {
		host, err := scanHost(rows)
		if err != nil {
			return nil, fmt.Errorf("scan host: %w", err)
		}
		hosts = append(hosts, *host)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate hosts: %w", err)
	}

	return hosts, nil
}

func (r *HostRepository) Update(
	ctx context.Context,
	host *models.Host,
) (*models.Host, error) {
	query := `
		UPDATE hosts
		SET
			name = $2,
			hostname = $3,
			description = $4,
			os = $5,
			architecture = $6,
			kernel = $7,
			updated_at = NOW()
		WHERE id = $1
		RETURNING ` + hostColumns

	row := r.db.QueryRow(
		ctx, query,
		host.ID,
		host.Name,
		host.Hostname,
		host.Description,
		host.OS,
		host.Architecture,
		host.Kernel,
	)

	result, err := scanHost(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrHostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update host: %w", err)
	}

	return result, nil
}

func (r *HostRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM hosts WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete host: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrHostNotFound
	}

	return nil
}
