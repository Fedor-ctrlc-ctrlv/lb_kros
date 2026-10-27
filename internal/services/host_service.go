package services

import (
	"context"
	"errors"
	"strings"

	"web-api/internal/models"
	"web-api/internal/repository"
)

var (
	ErrInvalidHost   = errors.New("invalid host data")
	ErrInvalidHostID = errors.New("invalid host id")
)

type HostStore interface {
	Create(context.Context, *models.Host) (*models.Host, error)
	GetByID(context.Context, int64) (*models.Host, error)
	List(context.Context) ([]models.Host, error)
	Update(context.Context, *models.Host) (*models.Host, error)
	Delete(context.Context, int64) error
}

type HostService struct {
	repo HostStore
}

func NewHostService(repo HostStore) *HostService {
	return &HostService{repo: repo}
}

type CreateHostInput struct {
	Name         string `json:"name"`
	Hostname     string `json:"hostname"`
	Description  string `json:"description"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Kernel       string `json:"kernel"`
}

type UpdateHostInput struct {
	Name         *string `json:"name"`
	Hostname     *string `json:"hostname"`
	Description  *string `json:"description"`
	OS           *string `json:"os"`
	Architecture *string `json:"architecture"`
	Kernel       *string `json:"kernel"`
}

func (s *HostService) Create(
	ctx context.Context,
	input CreateHostInput,
) (*models.Host, error) {
	host := &models.Host{
		Name:         strings.TrimSpace(input.Name),
		Hostname:     strings.TrimSpace(input.Hostname),
		Description:  strings.TrimSpace(input.Description),
		OS:           strings.TrimSpace(input.OS),
		Architecture: strings.TrimSpace(input.Architecture),
		Kernel:       strings.TrimSpace(input.Kernel),
	}

	if host.Name == "" || host.Hostname == "" {
		return nil, ErrInvalidHost
	}

	return s.repo.Create(ctx, host)
}

func (s *HostService) GetByID(
	ctx context.Context,
	id int64,
) (*models.Host, error) {
	if id <= 0 {
		return nil, ErrInvalidHostID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *HostService) List(
	ctx context.Context,
) ([]models.Host, error) {
	return s.repo.List(ctx)
}

func (s *HostService) Update(
	ctx context.Context,
	id int64,
	input UpdateHostInput,
) (*models.Host, error) {
	if id <= 0 {
		return nil, ErrInvalidHostID
	}

	host, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		host.Name = strings.TrimSpace(*input.Name)
	}
	if input.Hostname != nil {
		host.Hostname = strings.TrimSpace(*input.Hostname)
	}
	if input.Description != nil {
		host.Description = strings.TrimSpace(*input.Description)
	}
	if input.OS != nil {
		host.OS = strings.TrimSpace(*input.OS)
	}
	if input.Architecture != nil {
		host.Architecture = strings.TrimSpace(*input.Architecture)
	}
	if input.Kernel != nil {
		host.Kernel = strings.TrimSpace(*input.Kernel)
	}

	if host.Name == "" || host.Hostname == "" {
		return nil, ErrInvalidHost
	}

	return s.repo.Update(ctx, host)
}

func (s *HostService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrInvalidHostID
	}

	return s.repo.Delete(ctx, id)
}

// Используется обработчиком для определения ошибки 404.
var ErrHostNotFound = repository.ErrHostNotFound
