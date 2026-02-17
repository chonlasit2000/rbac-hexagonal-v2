package port

import (
	"context"
)

// Service Port (Use Case)
type AuthService interface {
	Register(ctx context.Context, req *RegisterReq) error
	Login(ctx context.Context, req *LoginReq) (*AuthResponse, error)
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
