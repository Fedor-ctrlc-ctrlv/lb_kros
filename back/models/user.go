package models

import (
	"errors"
	"strings"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u User) CanWrite() bool {
	return u.IsAdmin()
}

func (u *User) ChangeRole(role UserRole) error {
	if role != RoleUser && role != RoleAdmin {
		return errors.New("invalid user role")
	}

	u.Role = role
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("username cannot be empty")
	}

	if u.Role != RoleUser && u.Role != RoleAdmin {
		return errors.New("invalid user role")
	}

	return nil
}
