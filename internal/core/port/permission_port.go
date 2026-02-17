package port

import (
	"context"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
)

type PermissionRepository interface {
	Create(ctx context.Context, perm *domain.Permission) error
	GetAll(ctx context.Context) ([]domain.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*domain.Permission, error)
}
