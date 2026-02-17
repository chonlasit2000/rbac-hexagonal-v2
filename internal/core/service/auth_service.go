package service

import (
	"context"
	"errors"
	"time"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/port"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo  port.UserRepository
	jwtSecret string
}

func NewAuthService(repo port.UserRepository, secret string) port.AuthService {
	return &authService{userRepo: repo, jwtSecret: secret}
}

func (s *authService) Register(ctx context.Context, req *port.RegisterReq) error {
	// 1. Hash Password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return err
	}

	// 2. Prepare User
	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashed),
	}

	// 3. Save to Repo
	return s.userRepo.Create(ctx, user)
}

func (s *authService) Login(ctx context.Context, req *port.LoginReq) (*port.AuthResponse, error) {
	// 1. Find User
	user, err := s.userRepo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 2. Check Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. Generate JWT
	claims := jwt.MapClaims{
		"user_id":  user.Uid,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &port.AuthResponse{AccessToken: t}, nil
}
