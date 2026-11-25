package services

import (
	"context"
	"errors"
	"strings"

	"web-api/internal/models"
)

var ErrInvalidPolicy = errors.New("invalid policy data")

type PolicyStore interface {
	Create(
		context.Context,
		*models.SecurityPolicy,
	) (*models.SecurityPolicy, error)

	GetByID(
		context.Context,
		int64,
	) (*models.SecurityPolicy, error)

	List(context.Context) ([]models.SecurityPolicy, error)

	Update(
		context.Context,
		*models.SecurityPolicy,
	) (*models.SecurityPolicy, error)

	Delete(context.Context, int64) error
}

type PolicyService struct {
	repo PolicyStore
}

func NewPolicyService(repo PolicyStore) *PolicyService {
	return &PolicyService{repo: repo}
}

type CreatePolicyInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatePolicyInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (s *PolicyService) Create(
	ctx context.Context,
	input CreatePolicyInput,
) (*models.SecurityPolicy, error) {
	policy := &models.SecurityPolicy{
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
	}

	if err := policy.Validate(); err != nil {
		return nil, ErrInvalidPolicy
	}

	return s.repo.Create(ctx, policy)
}

func (s *PolicyService) GetByID(
	ctx context.Context,
	id int64,
) (*models.SecurityPolicy, error) {
	if id <= 0 {
		return nil, ErrInvalidPolicy
	}

	return s.repo.GetByID(ctx, id)
}

func (s *PolicyService) List(
	ctx context.Context,
) ([]models.SecurityPolicy, error) {
	return s.repo.List(ctx)
}

func (s *PolicyService) Update(
	ctx context.Context,
	id int64,
	input UpdatePolicyInput,
) (*models.SecurityPolicy, error) {
	if id <= 0 {
		return nil, ErrInvalidPolicy
	}

	policy, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		policy.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		policy.Description = strings.TrimSpace(*input.Description)
	}

	if err := policy.Validate(); err != nil {
		return nil, ErrInvalidPolicy
	}

	return s.repo.Update(ctx, policy)
}

func (s *PolicyService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrInvalidPolicy
	}

	return s.repo.Delete(ctx, id)
}
