package models

import "time"

type AuthUser struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         UserRole
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type CurrentUser struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Role     UserRole `json:"role"`
}
