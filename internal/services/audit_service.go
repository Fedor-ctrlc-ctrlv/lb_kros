package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"web-api/internal/models"
	"web-api/internal/repository"
)

var ErrInvalidSnapshot = errors.New("invalid snapshot")

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Import(
	ctx context.Context,
	hostID int64,
	raw []byte,
) (*repository.AuditRecord, error) {
	if hostID <= 0 {
		return nil, ErrInvalidHostID
	}

	var snapshot models.Snapshot
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSnapshot, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, ErrInvalidSnapshot
	}

	if err := validateSnapshot(&snapshot); err != nil {
		return nil, err
	}

	return s.repo.Import(ctx, hostID, &snapshot, json.RawMessage(raw))
}

func validateSnapshot(s *models.Snapshot) error {
	if s.Version != "2.0" ||
		s.CreatedAt.IsZero() ||
		s.CreatedAt.After(time.Now().Add(5*time.Minute)) ||
		strings.TrimSpace(s.System.Hostname) == "" {
		return ErrInvalidSnapshot
	}

	if len(s.Files) == 0 {
		return fmt.Errorf("%w: empty files collection", ErrInvalidSnapshot)
	}

	paths := make(map[string]bool, len(s.Files))
	for _, f := range s.Files {
		if f.Path == "" || paths[f.Path] {
			return fmt.Errorf("%w: duplicate or empty file path", ErrInvalidSnapshot)
		}
		if f.Type != "file" &&
			f.Type != "directory" &&
			f.Type != "symlink" {
			return fmt.Errorf("%w: unsupported file type %q",
				ErrInvalidSnapshot, f.Type)
		}
		paths[f.Path] = true
	}

	users := make(map[string]bool)
	for _, u := range s.Users {
		if u.Username == "" || users[u.Username] {
			return fmt.Errorf("%w: duplicate username", ErrInvalidSnapshot)
		}
		users[u.Username] = true
	}

	groups := make(map[string]bool)
	for _, g := range s.Groups {
		if g.Name == "" || groups[g.Name] {
			return fmt.Errorf("%w: duplicate group", ErrInvalidSnapshot)
		}
		groups[g.Name] = true
	}

	services := make(map[string]bool)
	for _, service := range s.Services {
		if service.Name == "" || service.Name == "●" ||
			services[service.Name] {
			return fmt.Errorf("%w: invalid or duplicate service name %q",
				ErrInvalidSnapshot, service.Name)
		}
		services[service.Name] = true
	}

	cronSources := make(map[string]bool)
	for _, job := range s.CronJobs {
		if job.Source == "" || cronSources[job.Source] {
			return fmt.Errorf("%w: duplicate cron source", ErrInvalidSnapshot)
		}
		cronSources[job.Source] = true
	}

	return nil
}

func (s *AuditService) GetByID(
	ctx context.Context, id int64,
) (*repository.AuditRecord, error) {
	if id <= 0 {
		return nil, ErrInvalidHostID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *AuditService) ListByHost(
	ctx context.Context, hostID int64,
) ([]repository.AuditRecord, error) {
	if hostID <= 0 {
		return nil, ErrInvalidHostID
	}
	return s.repo.ListByHost(ctx, hostID)
}

func (s *AuditService) GetSnapshot(
	ctx context.Context, id int64,
) (json.RawMessage, error) {
	if id <= 0 {
		return nil, ErrInvalidHostID
	}
	return s.repo.GetSnapshot(ctx, id)
}

func (s *AuditService) GetSummary(
	ctx context.Context, id int64,
) (*repository.AuditSummary, error) {
	if id <= 0 {
		return nil, ErrInvalidHostID
	}
	return s.repo.GetSummary(ctx, id)
}

func (s *AuditService) ListByHostPeriod(
	ctx context.Context,
	hostID int64,
	from, to time.Time,
) ([]repository.AuditRecord, error) {
	if hostID <= 0 {
		return nil, ErrInvalidHostID
	}

	if from.IsZero() || to.IsZero() || !from.Before(to) {
		return nil, fmt.Errorf("invalid time range")
	}

	return s.repo.ListByHostPeriod(ctx, hostID, from, to)
}
