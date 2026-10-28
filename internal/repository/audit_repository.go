package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"web-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{db: db}
}

type AuditRecord struct {
	ID                int64      `json:"id"`
	HostID            int64      `json:"host_id"`
	Status            string     `json:"status"`
	SnapshotVersion   string     `json:"snapshot_version"`
	SnapshotCreatedAt time.Time  `json:"snapshot_created_at"`
	StartedAt         time.Time  `json:"started_at"`
	FinishedAt        *time.Time `json:"finished_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

type AuditSummary struct {
	AuditID  int64 `json:"audit_id"`
	Files    int64 `json:"files"`
	Users    int64 `json:"users"`
	Groups   int64 `json:"groups"`
	Services int64 `json:"services"`
	CronJobs int64 `json:"cron_jobs"`
}

const auditColumns = `
	id, host_id, status, snapshot_version,
	snapshot_created_at, started_at, finished_at, created_at
`

func scanAudit(row pgx.Row) (*AuditRecord, error) {
	var a AuditRecord
	err := row.Scan(
		&a.ID, &a.HostID, &a.Status,
		&a.SnapshotVersion, &a.SnapshotCreatedAt,
		&a.StartedAt, &a.FinishedAt, &a.CreatedAt,
	)
	return &a, err
}

func (r *AuditRepository) Import(
	ctx context.Context,
	hostID int64,
	snapshot *models.Snapshot,
	raw json.RawMessage,
) (*AuditRecord, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var hostname string
	err = tx.QueryRow(ctx,
		`SELECT hostname FROM hosts WHERE id = $1 FOR UPDATE`,
		hostID,
	).Scan(&hostname)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrHostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find host: %w", err)
	}
	if hostname != snapshot.System.Hostname {
		return nil, models.ErrSnapshotHostMismatch
	}

	var audit AuditRecord
	err = tx.QueryRow(ctx, `
		INSERT INTO audits (
			host_id, status, snapshot_version,
			snapshot_created_at, raw_snapshot,
			started_at, finished_at
		)
		VALUES ($1, 'COMPLETED', $2, $3, $4, NOW(), NOW())
		RETURNING `+auditColumns,
		hostID, snapshot.Version, snapshot.CreatedAt, raw,
	).Scan(
		&audit.ID, &audit.HostID, &audit.Status,
		&audit.SnapshotVersion, &audit.SnapshotCreatedAt,
		&audit.StartedAt, &audit.FinishedAt, &audit.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert audit: %w", err)
	}

	if err := copyFiles(ctx, tx, audit.ID, snapshot.Files); err != nil {
		return nil, err
	}
	if err := copyUsers(ctx, tx, audit.ID, snapshot.Users); err != nil {
		return nil, err
	}
	if err := copyGroups(ctx, tx, audit.ID, snapshot.Groups); err != nil {
		return nil, err
	}
	if err := copyServices(ctx, tx, audit.ID, snapshot.Services); err != nil {
		return nil, err
	}
	if err := copyCron(ctx, tx, audit.ID, snapshot.CronJobs); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit audit: %w", err)
	}

	return &audit, nil
}

func copyFiles(
	ctx context.Context, tx pgx.Tx, auditID int64,
	files []models.SnapshotFile,
) error {
	rows := make([][]any, 0, len(files))
	for _, f := range files {
		rows = append(rows, []any{
			auditID, f.Path, f.Type, f.LinkTarget,
			f.SHA256, f.HashStatus, f.Permissions,
			f.Mode, f.UID, f.GID, f.Size, f.ModifiedAt,
			f.SUID, f.SGID, f.Sticky,
		})
	}
	return copyRows(ctx, tx, "fs_nodes", []string{
		"audit_id", "path", "node_type", "link_target",
		"sha256", "hash_status", "permissions", "mode",
		"uid", "gid", "size", "modified_at",
		"suid", "sgid", "sticky",
	}, rows)
}

func copyUsers(
	ctx context.Context, tx pgx.Tx, auditID int64,
	users []models.SnapshotUser,
) error {
	rows := make([][]any, 0, len(users))
	for _, u := range users {
		raw, err := json.Marshal(u)
		if err != nil {
			return err
		}
		rows = append(rows, []any{
			auditID, u.Username, u.UID, u.GID,
			u.Home, u.Shell, string(raw),
		})
	}
	return copyRows(ctx, tx, "linux_users", []string{
		"audit_id", "username", "uid", "gid",
		"home", "shell", "raw_data",
	}, rows)
}

func copyGroups(
	ctx context.Context, tx pgx.Tx, auditID int64,
	groups []models.SnapshotGroup,
) error {
	rows := make([][]any, 0, len(groups))
	for _, g := range groups {
		members := g.Members
		if members == nil {
			members = []string{}
		}
		rows = append(rows, []any{
			auditID, g.Name, g.GID, members,
		})
	}
	return copyRows(ctx, tx, "linux_groups", []string{
		"audit_id", "name", "gid", "members",
	}, rows)
}

func copyServices(
	ctx context.Context, tx pgx.Tx, auditID int64,
	services []models.SnapshotService,
) error {
	rows := make([][]any, 0, len(services))
	for _, s := range services {
		rows = append(rows, []any{
			auditID, s.Name, s.LoadState,
			s.ActiveState, s.SubState,
		})
	}
	return copyRows(ctx, tx, "services", []string{
		"audit_id", "name", "load_state",
		"active_state", "sub_state",
	}, rows)
}

func copyCron(
	ctx context.Context, tx pgx.Tx, auditID int64,
	jobs []models.SnapshotCron,
) error {
	rows := make([][]any, 0, len(jobs))
	for _, j := range jobs {
		rows = append(rows, []any{
			auditID, j.Source, j.Content,
		})
	}
	return copyRows(ctx, tx, "cron_jobs", []string{
		"audit_id", "source", "content",
	}, rows)
}

func copyRows(
	ctx context.Context,
	tx pgx.Tx,
	table string,
	columns []string,
	rows [][]any,
) error {
	if len(rows) == 0 {
		return nil
	}

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{table},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy %s: %w", table, err)
	}
	return nil
}

func (r *AuditRepository) GetByID(
	ctx context.Context, id int64,
) (*AuditRecord, error) {
	audit, err := scanAudit(r.db.QueryRow(ctx,
		`SELECT `+auditColumns+` FROM audits WHERE id = $1`,
		id,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrAuditNotFound
	}
	return audit, err
}

func (r *AuditRepository) ListByHost(
	ctx context.Context, hostID int64,
) ([]AuditRecord, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+auditColumns+`
		 FROM audits WHERE host_id = $1 ORDER BY id DESC`,
		hostID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]AuditRecord, 0)
	for rows.Next() {
		a, err := scanAudit(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *a)
	}
	return result, rows.Err()
}

func (r *AuditRepository) GetSnapshot(
	ctx context.Context, id int64,
) (json.RawMessage, error) {
	var raw json.RawMessage
	err := r.db.QueryRow(ctx,
		`SELECT raw_snapshot FROM audits WHERE id = $1`,
		id,
	).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrAuditNotFound
	}
	return raw, err
}

func (r *AuditRepository) GetSummary(
	ctx context.Context, id int64,
) (*AuditSummary, error) {
	if _, err := r.GetByID(ctx, id); err != nil {
		return nil, err
	}

	var s AuditSummary
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM fs_nodes WHERE audit_id = $1),
			(SELECT COUNT(*) FROM linux_users WHERE audit_id = $1),
			(SELECT COUNT(*) FROM linux_groups WHERE audit_id = $1),
			(SELECT COUNT(*) FROM services WHERE audit_id = $1),
			(SELECT COUNT(*) FROM cron_jobs WHERE audit_id = $1)
	`, id).Scan(
		&s.Files, &s.Users, &s.Groups,
		&s.Services, &s.CronJobs,
	)
	if err != nil {
		return nil, err
	}
	s.AuditID = id
	return &s, nil
}
