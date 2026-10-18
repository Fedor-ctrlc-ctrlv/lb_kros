package repository

import (
	"context"
	"errors"

	"web-api/back/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*models.AuthUser, error) {
	var user models.AuthUser

	err := r.db.QueryRow(ctx, `
		SELECT id, username, password_hash, role
		FROM app_users
		WHERE username = $1
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) GetByID(
	ctx context.Context,
	id int64,
) (*models.AuthUser, error) {
	var user models.AuthUser

	err := r.db.QueryRow(ctx, `
		SELECT id, username, password_hash, role
		FROM app_users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
