package services

import (
	"context"
	"errors"

	"web-api/back/repository"
)

var ErrInvalidBaseline = errors.New("invalid baseline data")

type BaselineStore interface {
	Set(context.Context, int64, int64) (*repository.AuditRecord, error)
	Get(context.Context, int64) (*repository.AuditRecord, error)
	Clear(context.Context, int64) error
}

type BaselineService struct {
	repo BaselineStore
}

func NewBaselineService(repo BaselineStore) *BaselineService {
	return &BaselineService{repo: repo}
}

type SetBaselineInput struct {
	AuditID int64 `json:"audit_id"`
}

func (s *BaselineService) Set(
	ctx context.Context,
	hostID int64,
	input SetBaselineInput,
) (*repository.AuditRecord, error) {
	if hostID <= 0 || input.AuditID <= 0 {
		return nil, ErrInvalidBaseline
	}

	return s.repo.Set(ctx, hostID, input.AuditID)
}

func (s *BaselineService) Get(
	ctx context.Context,
	hostID int64,
) (*repository.AuditRecord, error) {
	if hostID <= 0 {
		return nil, ErrInvalidBaseline
	}
	return s.repo.Get(ctx, hostID)
}

func (s *BaselineService) Clear(
	ctx context.Context,
	hostID int64,
) error {
	if hostID <= 0 {
		return ErrInvalidBaseline
	}
	return s.repo.Clear(ctx, hostID)
}
