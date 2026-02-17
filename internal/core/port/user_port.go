package port

import (
	"context"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetUserByUID(ctx context.Context, uid string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	AddAccosiateRole(ctx context.Context, userID string, roleID string) error
}
