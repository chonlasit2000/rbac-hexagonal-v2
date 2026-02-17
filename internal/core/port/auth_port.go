package port

import (
	"context"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
)

// Service Port (Use Case)
type AuthService interface {
	Register(ctx context.Context, req *RegisterReq) error
	Login(ctx context.Context, req *LoginReq) (*AuthResponse, error)
}

// Repository Port (Driven)
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
}

// --- DTOs (Request/Response) ---
type RegisterReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}
