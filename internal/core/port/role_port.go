package port

import (
	"context"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
)

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetAll(ctx context.Context) ([]domain.Role, error)
	GetRoleByUserUID(ctx context.Context, uid string) ([]domain.Role, error)
	GetRoleByName(ctx context.Context, name string) (*domain.Role, error)
	AddAccosiatePermission(ctx context.Context, roleID string, permID string) error
}

type RoleService interface {
	CreateRole(ctx context.Context, req *CreateRoleReq) error
}
