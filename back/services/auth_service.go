package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"web-api/back/models"
	"web-api/back/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthStore interface {
	GetByUsername(context.Context, string) (*models.AuthUser, error)
	GetByID(context.Context, int64) (*models.AuthUser, error)
}

type AuthService struct {
	repo   AuthStore
	secret []byte
}

func NewAuthService(repo AuthStore, secret string) *AuthService {
	return &AuthService{
		repo:   repo,
		secret: []byte(secret),
	}
}

func (s *AuthService) Login(
	ctx context.Context,
	input models.LoginInput,
) (*models.LoginResponse, error) {
	username := strings.TrimSpace(input.Username)

	if username == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(input.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	expiresAt := now.Add(time.Hour)

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(user.ID, 10),
		Issuer:    "linux-audit-api",
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *AuthService) Authenticate(
	ctx context.Context,
	tokenString string,
) (*models.CurrentUser, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}
			return s.secret, nil
		},
		jwt.WithIssuer("linux-audit-api"),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	if user.Role != models.RoleUser &&
		user.Role != models.RoleAdmin {
		return nil, ErrInvalidToken
	}

	return &models.CurrentUser{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}
